/*
  	dela - web TODO list
    Copyright (C) 2024, 2025  Kasyanov Nikolay Alexeyevich (Unbewohnte)

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

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
