package server

type Appstate struct {
	Hibernation      HibernationState  `json:"hibernation"`
	CalendarMappings map[string]string `json:"calendarMappings"`
}

type HibernationState struct {
	Hibernate bool `json:"hibernate"`
}
