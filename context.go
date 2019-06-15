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

package apcore

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const (
	targetUserUUIDContextKey     = "targetUserUUID"
	activityIRIContextKey        = "activityIRI"
	activityTypeContextKey       = "activityType"
	completeRequestURLContextKey = "completeRequestURL"
)

type ctx struct {
	context.Context
}

func newPostRequestContext(r *http.Request, db *database) (c ctx, err error) {
	c = ctx{r.Context()}
	// TODO
	c.WithTargetUserUUID("")
	c.WithCompleteRequestURL(r, "", "")
	c.WithActivityIRI(nil) // Optional
	c.WithActivityType("")
	return
}

func newGetRequestContext(r *http.Request, db *database) (c ctx, err error) {
	c = ctx{r.Context()}
	// TODO
	c.WithTargetUserUUID("")
	c.WithCompleteRequestURL(r, "", "")
	return
}

func (c ctx) WithTargetUserUUID(s string) {
	c.Context = context.WithValue(c.Context, targetUserUUIDContextKey, s)
}

func (c ctx) WithActivityIRI(u *url.URL) {
	c.Context = context.WithValue(c.Context, activityIRIContextKey, u)
}

func (c ctx) WithActivityType(s string) {
	c.Context = context.WithValue(c.Context, activityTypeContextKey, s)
}

func (c ctx) WithCompleteRequestURL(r *http.Request, scheme, host string) {
	u := *r.URL // Copy
	u.Host = host
	u.Scheme = scheme
	c.Context = context.WithValue(c.Context, completeRequestURLContextKey, &u)
}

func (c ctx) TargetUserUUID() (s string, err error) {
	v := c.Value(targetUserUUIDContextKey)
	var ok bool
	if v == nil {
		err = fmt.Errorf("no target user UUID in context")
	} else if s, ok = v.(string); !ok {
		err = fmt.Errorf("target user UUID in context is not a string")
	}
	return
}

func (c ctx) ActivityIRI() (u *url.URL, err error) {
	v := c.Value(activityIRIContextKey)
	var ok bool
	if v == nil {
		err = fmt.Errorf("no activity id in context")
	} else if u, ok = v.(*url.URL); !ok {
		err = fmt.Errorf("activity id in context is not a *url.URL")
	}
	return
}

func (c ctx) ActivityType() (s string, err error) {
	v := c.Value(activityTypeContextKey)
	var ok bool
	if v == nil {
		err = fmt.Errorf("no activity type in context")
	} else if s, ok = v.(string); !ok {
		err = fmt.Errorf("activity type in context is not a string")
	}
	return
}

func (c ctx) CompleteRequestURL() (u *url.URL, err error) {
	v := c.Value(completeRequestURLContextKey)
	var ok bool
	if v == nil {
		err = fmt.Errorf("no complete request URL in context")
	} else if u, ok = v.(*url.URL); !ok {
		err = fmt.Errorf("complete request URL in context is not a *url.URL")
	}
	return
}
