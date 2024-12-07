package models

import "time"

type Notification struct {
	Manager   string    `bson:"manager"`
	Message   string    `bson:"message"`
	Timestamp time.Time `bson:"timestamp"`
}
