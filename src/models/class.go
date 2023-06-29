package models

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Class struct {
	ShouldBeInMorning bool
	ShouldBeInAfternoon bool
	NeedsTeachers Teacher
	OnDays []Day
	ForGrade int
	Name string
	PeriodsPerWeek int
	ID primitive.ObjectID `bson:"-"`
	PriorityScore int
}
func (c *Class) AddToDB(db *mongo.Database, ctx context.Context) error {
	res, err := db.Collection("class").InsertOne(ctx, c,)
	if err != nil {
		return err
	}
	c.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}
