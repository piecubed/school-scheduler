package sourcing

import (
	"encoding/csv"
	"os"
	"scheduling/src/models"
	"scheduling/src/types"
	"strconv"
)

type classRow struct {
	Grade int
	Class string
	PeriodsPerWeek int
	Teacher string
	Morning bool
}

func ReadCsv(filename string) ([][]string, error) {

    // Open CSV file
    f, err := os.Open(filename)
    if err != nil {
        return [][]string{}, err
    }
    defer f.Close()

    // Read File into a Variable
    lines, err := csv.NewReader(f).ReadAll()
    if err != nil {
        return [][]string{}, err
    }

    return lines, nil
}

func ParseCSV(filePath string) (map[string]*models.Teacher, []*models.Class, []types.Grade) {
	lines, err := ReadCsv(filePath)
	if err != nil {
		panic(err)
	}
	classes := []*models.Class{}
	teachers := map[string]*models.Teacher{}
	grades := []types.Grade{}

	for _, line := range lines {
		g, err := strconv.Atoi(line[0])
		if err != nil {
			panic(err)
		}
		ppw, err := strconv.Atoi(line[2])
		if err != nil {
			panic(err)
		}

		data := classRow{
			Grade: g,
			Class: line[1],
			PeriodsPerWeek: ppw,
			Teacher: line[3],
			Morning: line[4] == "M",
		}

		// Add grade to list of grades
		gradeExists := false
		for _, grade := range grades {
			if data.Grade == int(grade) {
				gradeExists = true
			}
		}
		if !gradeExists {
			grades = append(grades, types.Grade(data.Grade))
		}

		// Add teacher to list of teachers, add to db
		var teacher *models.Teacher
		teacherExists := false
		for _, teach := range teachers {
			if data.Teacher == teach.TeacherName {
				teacherExists = true
				teacher = teach
			}
		}

		if !teacherExists {
			teacher = &models.Teacher{
				TeacherName: data.Teacher,
				ClassesTaught: []*models.Class{},
			}
			// database.InsertTeacher(teacher, db)
			teachers[teacher.TeacherName] = teacher
		}
		

		// Add class to list of classes, add to db
		classExists := false
		for _, class := range classes {
			if data.Class == class.Name && data.Grade == class.ForGrade {
				classExists = true
			}
		}
		if !classExists {
			score := data.PeriodsPerWeek
			if data.Morning {
				score += 20
			}

			class := &models.Class{
				NeedsTeachers: *teacher,
				Name: data.Class,
				ForGrade: data.Grade,
				PeriodsPerWeek: data.PeriodsPerWeek,
				PriorityScore: score,
			}
			classes = append(classes, class)
		}
	}

	return teachers, classes, grades
}

