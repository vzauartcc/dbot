package models

type StudentAssignment struct {
	StudentID     string `json:"studentId"`
	InstructorCID int    `json:"instructorCid"`
	Added         bool   `json:"added"`
}
