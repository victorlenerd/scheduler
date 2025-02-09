package project

import (
	_ "errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/hashicorp/go-hclog"
	"net/http"
	"scheduler0/pkg/constants"
	"scheduler0/pkg/fsm"
	"scheduler0/pkg/models"
	job_repo "scheduler0/pkg/repository/job"
	"scheduler0/pkg/scheduler0time"
	"scheduler0/pkg/utils"
	"time"
)

//go:generate mockery --name ProjectRepo --output ../mocks
type ProjectRepo interface {
	CreateOne(project *models.Project) (uint64, *utils.GenericError)
	GetOneByName(project *models.Project) *utils.GenericError
	GetOneByID(project *models.Project) *utils.GenericError
	List(offset uint64, limit uint64) ([]models.Project, *utils.GenericError)
	Count() (uint64, *utils.GenericError)
	UpdateOneByID(project models.Project) (uint64, *utils.GenericError)
	DeleteOneByID(project models.Project) (uint64, *utils.GenericError)
	GetBatchProjectsByIDs(projectIds []uint64) ([]models.Project, *utils.GenericError)
}

type projectRepo struct {
	fsmStore              fsm.Scheduler0RaftStore
	jobRepo               job_repo.JobRepo
	logger                hclog.Logger
	scheduler0RaftActions fsm.Scheduler0RaftActions
}

func NewProjectRepo(logger hclog.Logger, scheduler0RaftActions fsm.Scheduler0RaftActions, store fsm.Scheduler0RaftStore, jobRepo job_repo.JobRepo) ProjectRepo {
	return &projectRepo{
		fsmStore:              store,
		scheduler0RaftActions: scheduler0RaftActions,
		jobRepo:               jobRepo,
		logger:                logger.Named("project-repo"),
	}
}

// CreateOne creates a single project
func (projectRepo *projectRepo) CreateOne(project *models.Project) (uint64, *utils.GenericError) {
	if len(project.Name) < 1 {
		return 0, utils.HTTPGenericError(http.StatusBadRequest, "name field is required")
	}

	if len(project.Description) < 1 {
		return 0, utils.HTTPGenericError(http.StatusBadRequest, "description field is required")
	}

	projectWithName := models.Project{
		ID:   0,
		Name: project.Name,
	}

	_ = projectRepo.GetOneByName(project)
	if projectWithName.ID > 0 {
		return 0, utils.HTTPGenericError(http.StatusBadRequest, fmt.Sprintf("another project exist with the same name, project with id %v has the same name", projectWithName.ID))
	}
	schedulerTime := scheduler0time.GetSchedulerTime()
	now := schedulerTime.GetTime(time.Now())

	query, params, err := sq.Insert(constants.ProjectsTableName).
		Columns(
			constants.ProjectsNameColumn,
			constants.ProjectsDescriptionColumn,
			constants.ProjectsDateCreatedColumn,
		).
		Values(
			project.Name,
			project.Description,
			now,
		).ToSql()
	if err != nil {
		return 0, utils.HTTPGenericError(http.StatusInternalServerError, err.Error())
	}

	res, applyErr := projectRepo.scheduler0RaftActions.WriteCommandToRaftLog(projectRepo.fsmStore.GetRaft(), constants.CommandTypeDbExecute, query, params, []uint64{}, 0)
	if applyErr != nil {
		return 0, utils.HTTPGenericError(http.StatusInternalServerError, applyErr.Error())
	}

	if res == nil {
		return 0, utils.HTTPGenericError(http.StatusServiceUnavailable, "service is unavailable")
	}

	insertedId := res.Data.LastInsertedId
	project.ID = uint64(insertedId)

	getErr := projectRepo.GetOneByID(project)
	if getErr != nil {
		return 0, utils.HTTPGenericError(http.StatusInternalServerError, getErr.Error())
	}

	return uint64(insertedId), nil
}

// GetOneByName returns a project with a matching name
func (projectRepo *projectRepo) GetOneByName(project *models.Project) *utils.GenericError {
	projectRepo.fsmStore.GetDataStore().ConnectionLock()
	defer projectRepo.fsmStore.GetDataStore().ConnectionUnlock()

	selectBuilder := sq.Select(
		constants.ProjectsIdColumn,
		constants.ProjectsNameColumn,
		constants.ProjectsDescriptionColumn,
		constants.ProjectsDateCreatedColumn,
	).
		From(constants.ProjectsTableName).
		Where(fmt.Sprintf("%s = ?", constants.ProjectsNameColumn), project.Name).
		RunWith(projectRepo.fsmStore.GetDataStore().GetOpenConnection())

	rows, err := selectBuilder.Query()
	if err != nil {
		return utils.HTTPGenericError(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		err = rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.DateCreated,
		)
		if err != nil {
			return utils.HTTPGenericError(http.StatusInternalServerError, err.Error())
		}
		count += 1
	}
	if rows.Err() != nil {
		return utils.HTTPGenericError(http.StatusInternalServerError, err.Error())
	}

	if count == 0 {
		return utils.HTTPGenericError(http.StatusNotFound, "project with name : "+project.Name+" does not exist")
	}
	return nil
}

// GetOneByID returns a project that matches the uuid
func (projectRepo *projectRepo) GetOneByID(project *models.Project) *utils.GenericError {
	projectRepo.fsmStore.GetDataStore().ConnectionLock()
	defer projectRepo.fsmStore.GetDataStore().ConnectionUnlock()

	selectBuilder := sq.Select(
		constants.ProjectsIdColumn,
		constants.ProjectsNameColumn,
		constants.ProjectsDescriptionColumn,
		constants.ProjectsDateCreatedColumn,
	).
		From(constants.ProjectsTableName).
		Where(fmt.Sprintf("%s = ?", constants.ProjectsIdColumn), project.ID).
		RunWith(projectRepo.fsmStore.GetDataStore().GetOpenConnection())

	rows, err := selectBuilder.Query()
	if err != nil {
		return utils.HTTPGenericError(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		err = rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.DateCreated,
		)
		if err != nil {
			return utils.HTTPGenericError(http.StatusInternalServerError, err.Error())
		}
		count += 1
	}
	if rows.Err() != nil {
		return utils.HTTPGenericError(http.StatusInternalServerError, err.Error())
	}

	if count == 0 {
		return utils.HTTPGenericError(http.StatusNotFound, "project does not exist")
	}
	return nil
}

func (projectRepo *projectRepo) GetBatchProjectsByIDs(projectIds []uint64) ([]models.Project, *utils.GenericError) {
	projectRepo.fsmStore.GetDataStore().ConnectionLock()
	defer projectRepo.fsmStore.GetDataStore().ConnectionUnlock()

	if len(projectIds) < 1 {
		return []models.Project{}, nil
	}

	cachedProjectIds := map[uint64]uint64{}

	for _, projectId := range projectIds {
		if _, ok := cachedProjectIds[projectId]; !ok {
			cachedProjectIds[projectId] = projectId
		}
	}

	ids := []uint64{}
	for _, projectId := range cachedProjectIds {
		ids = append(ids, projectId)
	}

	projectIdsArgs := []interface{}{ids[0]}
	idParams := "?"

	i := 0
	for i < len(ids)-1 {
		idParams += ",?"
		i += 1
		projectIdsArgs = append(projectIdsArgs, ids[i])
	}

	selectBuilder := sq.Select(
		constants.ProjectsIdColumn,
		constants.ProjectsNameColumn,
		constants.ProjectsDescriptionColumn,
		constants.ProjectsDateCreatedColumn,
	).
		From(constants.ProjectsTableName).
		Where(fmt.Sprintf("%s in (%s)", constants.ProjectsIdColumn, idParams), projectIdsArgs...).
		RunWith(projectRepo.fsmStore.GetDataStore().GetOpenConnection())

	rows, err := selectBuilder.Query()
	defer rows.Close()
	if err != nil {
		return nil, utils.HTTPGenericError(http.StatusInternalServerError, err.Error())
	}
	count := 0
	projects := []models.Project{}
	for rows.Next() {
		project := models.Project{}
		err = rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.DateCreated,
		)
		if err != nil {
			return nil, utils.HTTPGenericError(http.StatusInternalServerError, err.Error())
		}
		projects = append(projects, project)
		count += 1
	}
	if rows.Err() != nil {
		return nil, utils.HTTPGenericError(http.StatusInternalServerError, err.Error())
	}

	return projects, nil
}

// List returns a paginated set of results
func (projectRepo *projectRepo) List(offset uint64, limit uint64) ([]models.Project, *utils.GenericError) {
	projectRepo.fsmStore.GetDataStore().ConnectionLock()
	defer projectRepo.fsmStore.GetDataStore().ConnectionUnlock()

	selectBuilder := sq.Select(
		constants.ProjectsIdColumn,
		constants.ProjectsNameColumn,
		constants.ProjectsDescriptionColumn,
		constants.ProjectsDateCreatedColumn,
	).
		From(constants.ProjectsTableName).
		Offset(offset).
		Limit(limit).
		RunWith(projectRepo.fsmStore.GetDataStore().GetOpenConnection())

	projects := []models.Project{}
	rows, err := selectBuilder.Query()
	defer rows.Close()
	if err != nil {
		return nil, utils.HTTPGenericError(http.StatusInternalServerError, err.Error())
	}
	for rows.Next() {
		project := models.Project{}
		err = rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.DateCreated,
		)
		if err != nil {
			return nil, utils.HTTPGenericError(http.StatusInternalServerError, err.Error())
		}
		projects = append(projects, project)
	}
	if rows.Err() != nil {
		return nil, utils.HTTPGenericError(http.StatusInternalServerError, rows.Err().Error())
	}

	return projects, nil
}

// Count return the number of projects
func (projectRepo *projectRepo) Count() (uint64, *utils.GenericError) {
	projectRepo.fsmStore.GetDataStore().ConnectionLock()
	defer projectRepo.fsmStore.GetDataStore().ConnectionUnlock()

	countQuery := sq.Select("count(*)").From(constants.ProjectsTableName).RunWith(projectRepo.fsmStore.GetDataStore().GetOpenConnection())
	rows, err := countQuery.Query()
	defer rows.Close()
	if err != nil {
		return 0, utils.HTTPGenericError(500, err.Error())
	}
	count := 0
	for rows.Next() {
		err = rows.Scan(
			&count,
		)
		if err != nil {
			return 0, utils.HTTPGenericError(500, err.Error())
		}
	}
	if err != nil {
		return 0, utils.HTTPGenericError(http.StatusInternalServerError, err.Error())
	}

	return uint64(count), nil
}

// UpdateOneByID updates a single project
func (projectRepo *projectRepo) UpdateOneByID(project models.Project) (uint64, *utils.GenericError) {
	updateQuery := sq.Update(constants.ProjectsTableName).
		Set(constants.ProjectsDescriptionColumn, project.Description).
		Where(fmt.Sprintf("%s = ?", constants.ProjectsIdColumn), project.ID)

	query, params, err := updateQuery.ToSql()
	if err != nil {
		return 0, utils.HTTPGenericError(http.StatusInternalServerError, err.Error())
	}

	res, applyErr := projectRepo.scheduler0RaftActions.WriteCommandToRaftLog(projectRepo.fsmStore.GetRaft(), constants.CommandTypeDbExecute, query, params, []uint64{}, 0)
	if applyErr != nil {
		return 0, utils.HTTPGenericError(http.StatusInternalServerError, applyErr.Error())
	}
	if res == nil {
		return 0, utils.HTTPGenericError(http.StatusServiceUnavailable, "service is unavailable")
	}

	count := res.Data.RowsAffected

	return uint64(count), nil
}

// DeleteOneByID deletes a single project
func (projectRepo *projectRepo) DeleteOneByID(project models.Project) (uint64, *utils.GenericError) {
	projectJobs, getAllErr := projectRepo.jobRepo.GetAllByProjectID(project.ID, 0, 1, "id")
	if getAllErr != nil {
		return 0, utils.HTTPGenericError(http.StatusInternalServerError, getAllErr.Error())
	}

	if len(projectJobs) > 0 {
		return 0, utils.HTTPGenericError(http.StatusBadRequest, "cannot delete project with jobs")
	}

	deleteQuery := sq.
		Delete(constants.ProjectsTableName).
		Where(fmt.Sprintf("%s = ?", constants.ProjectsIdColumn), project.ID)

	query, params, deleteErr := deleteQuery.ToSql()
	if deleteErr != nil {
		return 0, utils.HTTPGenericError(http.StatusInternalServerError, deleteErr.Error())
	}

	res, applyErr := projectRepo.scheduler0RaftActions.WriteCommandToRaftLog(projectRepo.fsmStore.GetRaft(), constants.CommandTypeDbExecute, query, params, []uint64{}, 0)
	if applyErr != nil {
		return 0, applyErr
	}
	if res == nil {
		return 0, utils.HTTPGenericError(http.StatusServiceUnavailable, "service is unavailable")
	}

	count := res.Data.RowsAffected

	return uint64(count), nil
}
