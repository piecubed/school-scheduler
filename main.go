package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"scheduling/src/models"
	"scheduling/src/sourcing"
	t "scheduling/src/types"
	"sort"
)

const (
	Monday    = 1
	Tuesday   = 2
	Wednesday = 3
	Thursday  = 4
	Friday    = 5
)

func DayToString(day t.DayOfWeek) string {
	if day == 1 {
		return "Monday"
	}
	if day == 2 {
		return "Tueday"
	}
	if day == 3 {
		return "Wednesday"
	}
	if day == 4 {
		return "Thursday"
	}
	if day == 5 {
		return "Friday"
	}
	return "You screwed up"
}

func MakeRawEmptySchedules(grades []t.Grade, teachers map[string]*models.Teacher) (t.SchoolTeacherSchedule, t.SchoolGradeSchedule) {
	gradeSchedule := t.SchoolGradeSchedule{}
	schoolSchedule := t.SchoolTeacherSchedule{}

	for _, grade := range grades {
		gradeSchedule[grade] = t.WeeklySchedule{
			Monday:    map[t.Period]*models.Class{1: nil, 2: nil, 3: nil, 4: nil, 5: nil, 6: nil, 7: nil, 8: nil},
			Tuesday:   map[t.Period]*models.Class{1: nil, 2: nil, 3: nil, 4: nil, 5: nil, 6: nil, 7: nil, 8: nil},
			Wednesday: map[t.Period]*models.Class{1: nil, 2: nil, 3: nil, 4: nil, 5: nil, 6: nil, 7: nil, 8: nil},
			Thursday:  map[t.Period]*models.Class{1: nil, 2: nil, 3: nil, 4: nil, 5: nil, 6: nil, 7: nil, 8: nil},
			Friday:    map[t.Period]*models.Class{1: nil, 2: nil, 3: nil, 4: nil, 5: nil, 6: nil, 7: nil, 8: nil},
		}
	}
	for _, teacher := range teachers {
		schoolSchedule[teacher.TeacherName] = t.WeeklySchedule{
			Monday:    map[t.Period]*models.Class{1: nil, 2: nil, 3: nil, 4: nil, 5: nil, 6: nil, 7: nil, 8: nil},
			Tuesday:   map[t.Period]*models.Class{1: nil, 2: nil, 3: nil, 4: nil, 5: nil, 6: nil, 7: nil, 8: nil},
			Wednesday: map[t.Period]*models.Class{1: nil, 2: nil, 3: nil, 4: nil, 5: nil, 6: nil, 7: nil, 8: nil},
			Thursday:  map[t.Period]*models.Class{1: nil, 2: nil, 3: nil, 4: nil, 5: nil, 6: nil, 7: nil, 8: nil},
			Friday:    map[t.Period]*models.Class{1: nil, 2: nil, 3: nil, 4: nil, 5: nil, 6: nil, 7: nil, 8: nil},
		}
		for _, class := range teacher.ClassesTaught {
			class.PriorityScore += len(teacher.ClassesTaught)
		}
	}
	return schoolSchedule, gradeSchedule
}

func CreateBlankSchedules() ([]*models.Class, t.SchoolTeacherSchedule, t.SchoolGradeSchedule, []t.Grade, map[string]*models.Teacher) {
	teachers, classes, grades := sourcing.ParseCSV("./classes.csv")

	classBuffer := []*models.Class{}
	for _, class := range classes {
		if class.PeriodsPerWeek > 5 {
			newClassPPW := class.PeriodsPerWeek - 5
			newClass := &models.Class{
				ShouldBeInMorning:   class.ShouldBeInMorning,
				ShouldBeInAfternoon: class.ShouldBeInAfternoon,
				NeedsTeachers:       class.NeedsTeachers,
				OnDays:              class.OnDays,
				ForGrade:            class.ForGrade,
				Name:                class.Name,
				PeriodsPerWeek:      newClassPPW,
				PriorityScore:       class.PriorityScore,
			}
			classBuffer = append(classBuffer, newClass)
			class.PeriodsPerWeek = 5
		}
	}

	classes = append(classes, classBuffer...)

	for _, class := range classes {
		teacher := class.NeedsTeachers.TeacherName
		teachers[teacher].ClassesTaught = append(teachers[teacher].ClassesTaught, class)
	}

	schoolSchedule, gradeSchedule := MakeRawEmptySchedules(grades, teachers)

	sort.SliceStable(classes, func(i, j int) bool {
		return classes[i].PriorityScore > classes[j].PriorityScore
	})
	return classes, schoolSchedule, gradeSchedule, grades, teachers
}
func FindSlotForWholeWeek(class *models.Class, schoolSchedule t.SchoolTeacherSchedule, gradeSchedule t.SchoolGradeSchedule) bool {
	teacher := class.NeedsTeachers.TeacherName
	availiblePeriods := FindCommonFreePeriodsThroughoutWeekForTeacherAndGrade(schoolSchedule[teacher], gradeSchedule[t.Grade(class.ForGrade)])

	slots := []t.SlotInfo{}
	for period, days := range availiblePeriods {
		if len(days) >= class.PeriodsPerWeek {

			sort.Slice(days, func(i, j int) bool {
				return days[i] < days[j]
			})
			isConsecutive := true
			for i := 0; i < len(days)-1; i++ {
				if days[i+1]-days[i] != 1 {
					isConsecutive = false
				}
			}
			slots = append(slots, t.SlotInfo{Days: days, IsConsecutive: isConsecutive, Period: period})
		}
	}

	sort.SliceStable(slots, func(i, j int) bool {
		return slots[j].IsConsecutive
	})
	if len(slots) == 0 {
		return FitClassIn(availiblePeriods, class, schoolSchedule, gradeSchedule, teacher)
	} else {
		periodForClass := slots[0].Period
		daysForClass := slots[0].Days[0:class.PeriodsPerWeek]
		class.OnDays = []models.Day{}

		for _, day := range daysForClass {
			if schoolSchedule[teacher][day][periodForClass] == nil {
				schoolSchedule[teacher][day][periodForClass] = class
				class.OnDays = append(class.OnDays, models.Day{Day: int(day), Periods: []models.Period{{Period: int(periodForClass)}}})
			} else {
				return false
			}
			if gradeSchedule[t.Grade(class.ForGrade)][day][periodForClass] == nil {
				gradeSchedule[t.Grade(class.ForGrade)][day][periodForClass] = class
			} else {
				return false
			}
		}
	}
	return true
}

func FitClassIn(commonPeriods map[t.Period][]t.DayOfWeek, class *models.Class, schoolSchedule t.SchoolTeacherSchedule, gradeSchedule t.SchoolGradeSchedule, teacher string) bool {
	newSlots := []t.SlotInfo{}
	for period, days := range commonPeriods {
		sort.Slice(days, func(i, j int) bool {
			return days[i] < days[j]
		})
		isConsecutive := true
		for i := 0; i < len(days)-1; i++ {
			if days[i+1]-days[i] != 1 {
				isConsecutive = false
			}
		}
		newSlots = append(newSlots, t.SlotInfo{Days: days, IsConsecutive: isConsecutive, Period: period})
	}

	if len(commonPeriods) == 0 {
		return false
	}

	sort.SliceStable(newSlots, func(i, j int) bool {
		return len(newSlots[i].Days) > len(newSlots[j].Days)
	})

	usedDays := newSlots[0].Days
	schedulingOn := []t.SlotInfo{newSlots[0]}
	morePeriodsRequired := class.PeriodsPerWeek - len(usedDays)

	sort.SliceStable(newSlots, func(i, j int) bool {
		return len(newSlots[i].Days) < len(newSlots[j].Days)
	})
	skippedSlots := []t.SlotInfo{}
	for n, slot := range newSlots {
		usableDays := []t.DayOfWeek{}
		for _, slotDay := range slot.Days {
			for _, usedDay := range usedDays {
				if slotDay != usedDay {
					usableDays = append(usableDays, slotDay)
				} else {
					skippedSlots = append(skippedSlots, t.SlotInfo{Days: []t.DayOfWeek{slotDay}, Period: slot.Period})
				}
			}
		}
		var daysUsing []t.DayOfWeek

		if len(usableDays) >= morePeriodsRequired {
			daysUsing = usableDays[0:morePeriodsRequired]
		} else {
			daysUsing = usableDays
		}

		schedulingOn = append(schedulingOn, t.SlotInfo{Days: daysUsing, Period: slot.Period})
		usedDays = append(usedDays, daysUsing...)
		morePeriodsRequired -= len(daysUsing)
		if morePeriodsRequired != 0 && n == len(newSlots)-1 {
			print()
		}
		if morePeriodsRequired == 0 {
			break
		}
	}
	if morePeriodsRequired != 0 {
		for i := range skippedSlots {
			j := rand.Intn(i + 1)
			skippedSlots[i], skippedSlots[j] = skippedSlots[j], skippedSlots[i]
		}

		for _, slot := range skippedSlots {
			schedulingOn = append(schedulingOn, slot)
			morePeriodsRequired--
			if morePeriodsRequired == 0 {
				break
			}
		}
	}

	if len(schedulingOn) == 0 {
		return false
	}
	class.OnDays = []models.Day{}
	for _, slot := range schedulingOn {
		for _, day := range slot.Days {
			if schoolSchedule[teacher][day][slot.Period] == nil {
				schoolSchedule[teacher][day][slot.Period] = class
				class.OnDays = append(class.OnDays, models.Day{Day: int(day), Periods: []models.Period{{Period: int(slot.Period)}}})
			} else {
				return false
			}
			if gradeSchedule[t.Grade(class.ForGrade)][day][slot.Period] == nil {
				gradeSchedule[t.Grade(class.ForGrade)][day][slot.Period] = class
			} else {
				return false
			}
		}
	}
	return true
}

func main() {
	classes, schoolSchedule, gradeSchedule, grades, teachers := CreateBlankSchedules()
	n := 0
	shouldRepeat := true
	for shouldRepeat {
		for _, class := range classes {
			shouldRepeat = !FindSlotForWholeWeek(class, schoolSchedule, gradeSchedule)
			if shouldRepeat {
				class.PriorityScore += 1
				break
			}
		}

		if shouldRepeat {
			for _, class := range classes {
				class.OnDays = []models.Day{}
			}
			n++
			schoolSchedule, gradeSchedule = MakeRawEmptySchedules(grades, teachers)
		} else {
			for _, class := range classes {
				shouldBe := class.PeriodsPerWeek
				for _, daily := range gradeSchedule[t.Grade(class.ForGrade)] {
					for _, scheduled := range daily {
						if class == scheduled {
							shouldBe--
						}
					}
				}
				if shouldBe != 0 {
					fmt.Printf("%v %v did not get scheduled enough.  %v periods scheduled, need %v\n", class.Name, class.ForGrade, class.PeriodsPerWeek-shouldBe, class.PeriodsPerWeek)
				}
			}
		}
	}
	fmt.Println(n)
	SaveInfo(schoolSchedule, gradeSchedule)
}

func SaveInfo(schoolSchedule t.SchoolTeacherSchedule, gradeSchedule t.SchoolGradeSchedule) {
	teachersCSV := [][]string{}

	for name, week := range schoolSchedule {

		teachersCSV = append(teachersCSV, []string{name, "1", "2", "3", "4", "5", "6", "7", "8"})
		for _, day := range []t.DayOfWeek{Monday, Tuesday, Wednesday, Thursday, Friday} {
			schedule := week[day]
			teachersCSV = append(teachersCSV, []string{DayToString(day), GetNameMaybeNil(schedule, 1), GetNameMaybeNil(schedule, 2), GetNameMaybeNil(schedule, 3), GetNameMaybeNil(schedule, 4), GetNameMaybeNil(schedule, 5), GetNameMaybeNil(schedule, 6), GetNameMaybeNil(schedule, 7), GetNameMaybeNil(schedule, 8)})
		}
	}

	f, err := os.Create("teacherScheduleOUT.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	writer.WriteAll(teachersCSV)

	gradesCSV := [][]string{}

	for _, grade := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12} {
		week := gradeSchedule[t.Grade(grade)]
		gradesCSV = append(gradesCSV, []string{fmt.Sprint(grade), "1", "2", "3", "4", "5", "6", "7", "8"})
		for _, day := range []t.DayOfWeek{Monday, Tuesday, Wednesday, Thursday, Friday} {
			schedule := week[day]
			gradesCSV = append(gradesCSV, []string{DayToString(day), GetNameMaybeNil(schedule, 1), GetNameMaybeNil(schedule, 2), GetNameMaybeNil(schedule, 3), GetNameMaybeNil(schedule, 4), GetNameMaybeNil(schedule, 5), GetNameMaybeNil(schedule, 6), GetNameMaybeNil(schedule, 7), GetNameMaybeNil(schedule, 8)})
		}
	}

	f2, err := os.Create("gradesScheduleOUT.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	writer2 := csv.NewWriter(f2)
	writer2.WriteAll(gradesCSV)
}

func GetNameMaybeNil(schedule t.DailySchedule, period int) string {
	if class, ok := schedule[t.Period(period)]; ok && class != nil {
		return class.Name
	}
	return ""
}

func FindFreePeriodsOnDay(teacherSchedule t.WeeklySchedule, day t.DayOfWeek) []t.Period {
	targetDay := teacherSchedule[day]
	freePeriods := []t.Period{}
	for k, v := range targetDay {
		if v == nil {
			freePeriods = append(freePeriods, t.Period(k))
		}
	}
	return freePeriods
}

func FindFreePeriodsForWeek(weeklySchedule t.WeeklySchedule) map[t.DayOfWeek][]t.Period {
	freePeriodsThroughWeek := map[t.DayOfWeek][]t.Period{}
	for i := 1; i < 6; i++ {
		freePeriodsThroughWeek[t.DayOfWeek(i)] = FindFreePeriodsOnDay(weeklySchedule, t.DayOfWeek(i))
	}
	return freePeriodsThroughWeek
}

func FindFreePeriodsForTeacherAndGrade(teacherSchedule t.WeeklySchedule, gradeSchedule t.WeeklySchedule, day t.DayOfWeek) []t.Period {
	teacherFreePeriods := FindFreePeriodsOnDay(teacherSchedule, day)
	gradeFreePeriods := FindFreePeriodsOnDay(gradeSchedule, day)

	common := []t.Period{}

	for _, teacherPeriod := range teacherFreePeriods {
		for _, gradePeriod := range gradeFreePeriods {
			if teacherPeriod == gradePeriod {
				common = append(common, teacherPeriod)
			}
		}
	}
	return common
}

func FindCommonFreePeriodsThroughoutWeekForTeacherAndGrade(teacherSchedule t.WeeklySchedule, gradeSchedule t.WeeklySchedule) map[t.Period][]t.DayOfWeek {
	teacherFreePeriods := FindFreePeriodsForWeek(teacherSchedule)
	gradeFreePeriods := FindFreePeriodsForWeek(gradeSchedule)
	teacherFreePeriodsByPeriod := map[t.Period][]t.DayOfWeek{}
	gradeFreePeriodsByPeriod := map[t.Period][]t.DayOfWeek{}

	for day, periods := range teacherFreePeriods {
		for _, period := range periods {
			if _, ok := teacherFreePeriodsByPeriod[period]; !ok {
				teacherFreePeriodsByPeriod[period] = []t.DayOfWeek{day}
			} else {
				teacherFreePeriodsByPeriod[period] = append(teacherFreePeriodsByPeriod[period], day)
			}
		}
	}

	for day, periods := range gradeFreePeriods {
		for _, period := range periods {
			if _, ok := gradeFreePeriodsByPeriod[period]; !ok {
				gradeFreePeriodsByPeriod[period] = []t.DayOfWeek{day}
			} else {
				gradeFreePeriodsByPeriod[period] = append(gradeFreePeriodsByPeriod[period], day)
			}
		}
	}
	common := map[t.Period][]t.DayOfWeek{}

	for period, teacherDays := range teacherFreePeriodsByPeriod {
		gradeDays := gradeFreePeriodsByPeriod[period]
		for _, teacherDay := range teacherDays {
			for _, gradeDay := range gradeDays {
				if teacherDay == gradeDay {
					if _, ok := common[period]; !ok {
						common[period] = []t.DayOfWeek{}
					}
					common[period] = append(common[period], teacherDay)
				}
			}
		}
	}
	return common
}
