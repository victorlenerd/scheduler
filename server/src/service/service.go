package service

import (
	"context"
	"scheduler0/server/src/utils"
)

type Service struct {
	Pool *utils.Pool
	Ctx  context.Context
}
