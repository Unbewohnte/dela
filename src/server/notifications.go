package server

import (
	"Unbewohnte/dela/db"
	"Unbewohnte/dela/email"
	"Unbewohnte/dela/logger"
	"fmt"
	"time"
)

type Notification struct {
	UserEmail string
	ToDo      db.Todo
}

func (s *Server) SendTODOSNotification(userEmail string, todos []*db.Todo) error {
	var err error
	switch len(todos) {
	case 0:
		err = nil

	case 1:
		err = s.emailer.SendEmail(
			email.NewEmail(
				s.config.Verification.Emailer.User,
				"Dela: TODO Notification",
				fmt.Sprintf("<p>Notifying you on your \"%s\" TODO.</p><p>Due date is %s</p>", todos[0].Text, todos[0].Due),
				[]string{userEmail},
			),
		)

	default:
		err = s.emailer.SendEmail(
			email.NewEmail(
				s.config.Verification.Emailer.User,
				"Dela: TODO Notification",
				fmt.Sprintf("<p>Notifying you on your \"%s\" TODO.</p><p>Due date is %s</p><p>There are also %d other TODOs nearing Due date.</p>", todos[0].Text, todos[0].Due, len(todos)-1),
				[]string{userEmail},
			),
		)
	}

	return err
}

func (s *Server) NotifyUserOnTodos(userEmail string) error {
	user, err := s.db.GetUser(userEmail)
	if err != nil {
		return err
	}

	if !user.NotifyOnTodos {
		return nil
	}

	todosDue, err := s.db.GetUserTodosDue(userEmail, uint64(time.Duration(time.Hour*24).Seconds()))
	if err != nil {
		return err
	}

	if len(todosDue) == 0 {
		return nil
	}

	logger.Info("[Server][Notifications Routine] Notifying %s with %d TODOs...", userEmail, len(todosDue))
	err = s.SendTODOSNotification(userEmail, todosDue)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) StartNotificationsRoutine(delay time.Duration) {
	logger.Info("[Server][Notifications Routine] Notifications Routine Started!")

	var failed bool = false
	for {
		logger.Info("[Server][Notifications Routine] Retrieving list of users to be notified...")
		users, err := s.db.GetAllUsersWithNotificationsOn()
		if err != nil {
			logger.Error("[Server][Notifications Routine] Failed to retrieve users with notification on: %s", err)
			failed = true
		}

		if !failed {
			for _, user := range users {
				err = s.NotifyUserOnTodos(user.Email)
				if err != nil {
					logger.Error("[Server][Notifications routine] Failed to notify %s: %s", user.Email, err)
					continue
				}
			}
		}

		failed = false
		time.Sleep(delay)
	}
}
