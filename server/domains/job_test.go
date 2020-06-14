package domains

import (
	"context"
	"cron-server/server/db"
	"testing"
	"time"
)

var (
	project  = Project{Name: "test project", Description: "test project"}
	jobOne   = Job{}
	jobTwo   = Job{}
	jobThree = Job{}
	jobCtx   = context.Background()
)

func TestJob_CreateOne(t *testing.T) {
	var jobsPool, _ = db.NewPool(db.CreateConnection, 5)
	defer jobsPool.Close()

	t.Log("Creating job returns error if required inbound fields are nil")
	{
		jobOne.CallbackUrl = "http://test-url"
		jobOne.Data = "some-data"
		jobOne.ProjectId = ""
		jobOne.CronSpec = "* * * * *"

		_, err := jobOne.CreateOne(jobsPool, jobCtx)

		if err == nil {
			t.Fatalf("\t\t  Model should require values")
		}
	}

	t.Log("Creating job returns error if project id does not exist")
	{
		jobTwo.CallbackUrl = "http://test-url"
		jobTwo.Data = "some-data"
		jobTwo.ProjectId = "test-project-id"
		jobTwo.StartDate = time.Now().Add(600000 * time.Second)
		jobTwo.CronSpec = "* * * * *"

		id, err := jobTwo.CreateOne(jobsPool, jobCtx)

		if err == nil {
			t.Fatalf("\t\t  Invalid project id does not exist but job with %v was created", id)
		}
	}

	t.Log("Creating job returns new id")
	{
		id, err := project.CreateOne(jobsPool, jobCtx)
		if err != nil {
			t.Fatalf("\t\t  Cannot create project %v", err)
		}

		if len(id) < 1 {
			t.Fatalf("\t\t  Project id is invalid %v", id)
		}

		project.ID = id
		jobThree.CallbackUrl = "http://test-url"
		jobThree.Data = "some-data"
		jobThree.ProjectId = id
		jobThree.StartDate = time.Now().Add(600000 * time.Second)
		jobThree.CronSpec = "* * * * *"

		_, err = jobThree.CreateOne(jobsPool, jobCtx)
		if err != nil {
			t.Fatalf("\t\t  Could not create job %v", err)
		}

		rowsAffected, err := jobThree.DeleteOne(jobsPool, jobCtx)
		if err != nil {
			t.Fatalf("\t\t Could not delete job %v", err)
		}

		rowsAffected, err = project.DeleteOne(jobsPool, jobCtx)
		if err != nil && rowsAffected < 1 {
			t.Fatalf("\t\t  Could not delete project %v", err)
		}
	}
}

func TestJob_UpdateOne(t *testing.T) {
	var jobsPool, _ = db.NewPool(db.CreateConnection, 5)
	defer jobsPool.Close()

	t.Log("Cannot update cron spec on job")
	{
		id, err := project.CreateOne(jobsPool, jobCtx)
		if err != nil {
			t.Fatalf("\t\t  Cannot create project %v", err)
		}

		if len(id) < 1 {
			t.Fatalf("\t\t  Project id is invalid %v", id)
		}

		jobThree.ProjectId = id
		jobThree.CronSpec = "1 * * * *"

		id, err = jobThree.CreateOne(jobsPool, jobCtx)
		if err != nil {
			t.Fatalf("\t\t Could not update job %v", err)
		}

		jobThree.CronSpec = "2 * * * *"
		_, err = jobThree.UpdateOne(jobsPool, jobCtx)
		if err == nil {
			t.Fatalf("\t\t Could not update job %v", err)
		}

		jobThreePlaceholder := Job{ID: jobThree.ID}
		_, err = jobThreePlaceholder.GetOne(jobsPool, jobCtx, "id = ?", jobThree.ID)
		if err != nil {
			t.Fatalf("\t\t Could not get job %v", err)
		}

		if jobThreePlaceholder.CronSpec == jobThree.CronSpec {
			t.Fatalf("\t\t CronSpec should be immutable")
		}

		_, err = jobThree.DeleteOne(jobsPool, jobCtx)
		if err != nil {
			t.Fatalf("\t\t Could not update job %v", err)
		}

		_, err = project.DeleteOne(jobsPool, jobCtx)
		if err != nil {
			t.Fatalf("\t\t Could not update job %v", err)
		}
	}
}

func TestJob_DeleteOne(t *testing.T) {
	var jobsPool, _ = db.NewPool(db.CreateConnection, 5)
	defer jobsPool.Close()

	t.Log("Delete jobs")
	{
		rowsAffected, err := jobThree.DeleteOne(jobsPool, jobCtx)
		if err != nil && rowsAffected > 0 {
			t.Fatalf("\t\t %v", err)
		}

		rowsAffected, err = project.DeleteOne(jobsPool, jobCtx)
		if err != nil && rowsAffected > 0 {
			t.Fatalf("\t\t %v", err)
		}
	}
}
