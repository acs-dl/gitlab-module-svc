package gitlab

import (
	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/helpers"
	"github.com/acs-dl/gitlab-module-svc/internal/pqueue"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func GetPermissionWithType(queue *pqueue.PriorityQueue, function any, args []any, priority int) (*TypeSub, error) {
	item, err := helpers.AddFunctionInPQueue(queue, function, args, priority)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add function in pqueue")
	}

	err = item.Response.Error
	if err != nil {
		return nil, errors.Wrap(err, "some error while getting permission type")
	}

	typeSub, ok := item.Response.Value.(*TypeSub)
	if !ok {
		return nil, errors.New("wrong response type")
	}

	return typeSub, nil
}

func GetUser(queue *pqueue.PriorityQueue, function any, args []any, priority int) (*data.User, error) {
	item, err := helpers.AddFunctionInPQueue(queue, function, args, priority)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add function in pqueue")
	}

	err = item.Response.Error
	if err != nil {
		return nil, errors.Wrap(err, "some error while getting user from api")
	}

	user, ok := item.Response.Value.(*data.User)
	if !ok {
		return nil, errors.New("wrong response type")
	}

	return user, nil
}

func GetUsers(queue *pqueue.PriorityQueue, function any, args []any, priority int) ([]data.User, error) {
	item, err := helpers.AddFunctionInPQueue(queue, function, args, priority)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add function in pqueue")
	}

	err = item.Response.Error
	if err != nil {
		return nil, errors.Wrap(err, "some error while getting chat users from api")
	}

	users, ok := item.Response.Value.([]data.User)
	if !ok {
		return nil, errors.New("wrong response type")
	}

	return users, nil
}

func GetPermissions(queue *pqueue.PriorityQueue, function any, args []any, priority int) ([]data.Permission, error) {
	item, err := helpers.AddFunctionInPQueue(queue, function, args, priority)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add function in pqueue")
	}

	err = item.Response.Error
	if err != nil {
		return nil, errors.Wrap(err, "some error while getting permissions from api")
	}

	permissions, ok := item.Response.Value.([]data.Permission)
	if !ok {
		return nil, errors.New("wrong response type")
	}

	return permissions, nil
}

func GetPermission(queue *pqueue.PriorityQueue, function any, args []any, priority int) (*data.Permission, error) {
	item, err := helpers.AddFunctionInPQueue(queue, function, args, priority)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add function in pqueue")
	}

	err = item.Response.Error
	if err != nil {
		return nil, errors.Wrap(err, "some error while getting permission from api")
	}

	permission, ok := item.Response.Value.(*data.Permission)
	if !ok {
		return nil, errors.New("wrong response type")
	}

	return permission, nil
}

func GetRequestError(queue *pqueue.PriorityQueue, function any, args []any, priority int) error {
	item, err := helpers.AddFunctionInPQueue(queue, function, args, priority)
	if err != nil {
		return errors.Wrap(err, "failed to add function in pqueue")
	}

	err = item.Response.Error
	if err != nil {
		return errors.Wrap(err, "some error while making request from api")
	}

	return nil
}
