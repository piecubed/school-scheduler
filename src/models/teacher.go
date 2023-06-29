package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Teacher struct {
	TeacherName string
	ClassesTaught []*Class;
	ID primitive.ObjectID `bson:"-"`
}