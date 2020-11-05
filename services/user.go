// apcore is a server framework for implementing an ActivityPub application.
// Copyright (C) 2019 Cory Slep
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package services

import (
	"database/sql"
	"net/url"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/models"
	"github.com/go-fed/apcore/paths"
	"github.com/go-fed/apcore/util"
)

// CreateUserParameters contains all parameters needed to create a user & Actor.
type CreateUserParameters struct {
	Scheme     string
	Host       string
	Username   string
	Email      string
	HashParams HashPasswordParameters
	RSAKeySize int
}

type Users struct {
	App         app.Application
	DB          *sql.DB
	Users       *models.Users
	PrivateKeys *models.PrivateKeys
	Inboxes     *models.Inboxes
	Outboxes    *models.Outboxes
	Followers   *models.Followers
	Following   *models.Following
	Liked       *models.Liked
}

func (u *Users) CreateUser(c util.Context, params CreateUserParameters, password string) (userID string, err error) {
	return u.createUser(c,
		params,
		password,
		models.Privileges{},
		models.Preferences{})
}

func (u *Users) CreateAdminUser(c util.Context, params CreateUserParameters, password string) (userID string, err error) {
	return u.createUser(c,
		params,
		password,
		models.Privileges{},
		models.Preferences{})
}

func (u *Users) createUser(c util.Context, params CreateUserParameters, password string, roles models.Privileges, prefs models.Preferences) (userID string, err error) {
	// Prepare Salt & Hashed Password
	var salt, hashpass []byte
	salt, hashpass, err = hashPass(params.HashParams, password)
	if err != nil {
		return
	}
	// Prepare PrivateKey
	var privKey []byte
	var pubKey string
	privKey, pubKey, err = createAndSerializeRSAKeys(params.RSAKeySize)
	if err != nil {
		return
	}

	return userID, doInTx(c, u.DB, func(tx *sql.Tx) error {
		// Insert into users table
		cu := &models.CreateUser{
			Email:       params.Email,
			Hashpass:    hashpass,
			Salt:        salt,
			Actor:       models.ActivityStreamsPerson{streams.NewActivityStreamsPerson()}, // Placeholder
			Privileges:  roles,
			Preferences: prefs,
		}
		userID, err = u.Users.Create(c, tx, cu)
		if err != nil {
			return err
		}
		// Create the ActivityStreams collections based on the userID.
		actor, actorID := toPersonActor(u.App,
			paths.UUID(userID),
			params.Scheme,
			params.Host,
			params.Username,
			params.Username, // preferredUsername
			"",              // summary
			pubKey)
		var inbox, outbox vocab.ActivityStreamsOrderedCollection
		inbox, err = emptyInbox(actorID)
		if err != nil {
			return err
		}
		outbox, err = emptyOutbox(actorID)
		if err != nil {
			return err
		}
		var followers, following, liked vocab.ActivityStreamsCollection
		followers, err = emptyFollowers(actorID)
		if err != nil {
			return err
		}
		following, err = emptyFollowing(actorID)
		if err != nil {
			return err
		}
		liked, err = emptyLiked(actorID)
		if err != nil {
			return err
		}
		// Update the created user with the filled-in actor
		err = u.Users.UpdateActor(c, tx, userID, models.ActivityStreamsPerson{actor})
		if err != nil {
			return err
		}
		// Insert into private_keys table
		err = u.PrivateKeys.Create(c, tx, userID, pKeyHttpSigPurpose, privKey)
		if err != nil {
			return err
		}
		// Insert empty inbox, outbox, followers, following, liked
		err = u.Inboxes.Create(c, tx, actorID, models.ActivityStreamsOrderedCollection{inbox})
		if err != nil {
			return err
		}
		err = u.Outboxes.Create(c, tx, actorID, models.ActivityStreamsOrderedCollection{outbox})
		if err != nil {
			return err
		}
		err = u.Followers.Create(c, tx, actorID, models.ActivityStreamsCollection{followers})
		if err != nil {
			return err
		}
		err = u.Following.Create(c, tx, actorID, models.ActivityStreamsCollection{following})
		if err != nil {
			return err
		}
		return u.Liked.Create(c, tx, actorID, models.ActivityStreamsCollection{liked})
	})
}

func (u *Users) ActorIDForOutbox(c util.Context, outboxIRI *url.URL) (actorIRI *url.URL, err error) {
	return actorIRI, doInTx(c, u.DB, func(tx *sql.Tx) error {
		var a models.URL
		a, err = u.Users.ActorIDForOutbox(c, tx, outboxIRI)
		if err != nil {
			return err
		}
		actorIRI = a.URL
		return nil
	})
}

func (u *Users) ActorIDForInbox(c util.Context, inboxIRI *url.URL) (actorIRI *url.URL, err error) {
	return actorIRI, doInTx(c, u.DB, func(tx *sql.Tx) error {
		var a models.URL
		a, err = u.Users.ActorIDForInbox(c, tx, inboxIRI)
		if err != nil {
			return err
		}
		actorIRI = a.URL
		return nil
	})
}