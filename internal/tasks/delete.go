package tasks

import (
	"context"

	"github.com/wellywell/shorturl/internal/storage"
	"go.uber.org/zap"
)

type Storage interface {
	DeleteBatch(ctx context.Context, records ...storage.ToDelete) error
}

func DeleteWorker(tasks <-chan storage.ToDelete, store Storage) {

	logger, err := zap.NewDevelopment()
	if err != nil {
		return
	}
	defer logger.Sync()

	sugar := logger.Sugar()

	sugar.Infoln("Started delete worker...")

	flushSize := 100

	var buffer []storage.ToDelete

	for {
		select {
		case task := <-tasks:
			sugar.Infof("Got record to delete %v", task)
			buffer = append(buffer, task)

			if len(buffer) == flushSize {
				err := store.DeleteBatch(context.Background(), buffer...)
				if err != nil {
					sugar.Error(err.Error())
				}
				buffer = nil
			}
		default:
			// новых задач нет, выполним текущие
			if len(buffer) > 0 {
				err := store.DeleteBatch(context.Background(), buffer...)
				if err != nil {
					sugar.Error(err.Error())
				}
				buffer = nil
			}
		}
	}
}
