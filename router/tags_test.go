package router

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestPostUserTags(t *testing.T) {
	e, cookie, mw := beforeTest(t)
	assert := assert.New(t)
	tagText := "post test"

	// 正常系
	post := struct {
		Tag string `json:"tag"`
	}{
		Tag: tagText,
	}
	body, err := json.Marshal(post)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "http://test", bytes.NewReader(body))
	c, rec := getContext(e, t, cookie, req)
	c.SetPath("/users/:userID/tags")
	c.SetParamNames("userID")
	c.SetParamValues(testUser.ID)
	requestWithContext(t, mw(PostUserTag), c)

	if assert.EqualValues(http.StatusCreated, rec.Code) {
		var responseBody []*TagForResponse
		if assert.NoError(json.Unmarshal(rec.Body.Bytes(), &responseBody)) {
			assert.Equal(tagText, responseBody[0].Tag)
			assert.NotEqual("", responseBody[0].ID)
		}
	}
}

func TestGetUserTags(t *testing.T) {
	e, cookie, mw := beforeTest(t)
	assert := assert.New(t)
	for i := 0; i < 5; i++ {
		mustMakeTag(t, testUser.ID, "tag"+strconv.Itoa(i))
	}

	// 正常系
	c, rec := getContext(e, t, cookie, nil)
	c.SetPath("/users/:userID/tags/")
	c.SetParamNames("userID")
	c.SetParamValues(testUser.ID)
	requestWithContext(t, mw(GetUserTags), c)

	if assert.EqualValues(http.StatusOK, rec.Code) {
		var responseBody []TagForResponse
		if assert.NoError(json.Unmarshal(rec.Body.Bytes(), &responseBody)) {
			assert.Len(responseBody, 5)
		}
	}
}

func TestPutUserTags(t *testing.T) {
	e, cookie, mw := beforeTest(t)
	assert := assert.New(t)
	tagText := "put test"

	// 正常系
	tag := mustMakeTag(t, testUser.ID, tagText)
	post := struct {
		IsLocked bool `json:"isLocked"`
	}{
		IsLocked: true,
	}
	body, err := json.Marshal(post)
	require.NoError(t, err)

	req := httptest.NewRequest("PUT", "http://test", bytes.NewReader(body))
	c, rec := getContext(e, t, cookie, req)
	c.SetPath("/users/:userID/tags/:tagID")
	c.SetParamNames("userID", "tagID")
	c.SetParamValues(testUser.ID, tag.TagID)
	requestWithContext(t, mw(PutUserTag), c)

	if assert.EqualValues(http.StatusOK, rec.Code) {
		var responseBody []*TagForResponse
		if assert.NoError(json.Unmarshal(rec.Body.Bytes(), &responseBody)) {
			assert.True(responseBody[0].IsLocked)
		}
	}
}

func TestDeleteUserTags(t *testing.T) {
	e, cookie, mw := beforeTest(t)
	tagText := "Delete test"

	// 正常系
	tag := mustMakeTag(t, testUser.ID, tagText)

	c, rec := getContext(e, t, cookie, nil)
	c.SetPath("/users/:userID/tags/:tagID")
	c.SetParamNames("userID", "tagID")
	c.SetParamValues(testUser.ID, tag.TagID)
	requestWithContext(t, mw(DeleteUserTag), c)

	assert.EqualValues(t, http.StatusNoContent, rec.Code)
}