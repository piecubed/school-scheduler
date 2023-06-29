package types

import "scheduling/src/models"

type SchoolTeacherSchedule map[string]WeeklySchedule
type WeeklySchedule map[DayOfWeek]DailySchedule
type DailySchedule map[Period]*models.Class
type DayOfWeek int

type Period int

type SchoolGradeSchedule map[Grade]WeeklySchedule
type Grade int
type SlotInfo struct {
	Days          []DayOfWeek
	IsConsecutive bool
	Period        Period
}