package router

import (
	"encoding/base64"
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"github.com/traPtitech/traQ/model"
	"github.com/traPtitech/traQ/oauth2"
	"github.com/traPtitech/traQ/oauth2/scope"
	"net/http"
	"regexp"
	"time"
)

var uriRegex = regexp.MustCompile(`^([a-z0-9+.-]+):(?://(?:((?:[a-z0-9-._~!$&'()*+,;=:]|%[0-9A-F]{2})*)@)?((?:[a-z0-9-._~!$&'()*+,;=]|%[0-9A-F]{2})*)(?::(\d*))?(/(?:[a-z0-9-._~!$&'()*+,;=:@/]|%[0-9A-F]{2})*)?|(/?(?:[a-z0-9-._~!$&'()*+,;=:@]|%[0-9A-F]{2})+(?:[a-z0-9-._~!$&'()*+,;=:@/]|%[0-9A-F]{2})*)?)(?:\?((?:[a-z0-9-._~!$&'()*+,;=:/?@]|%[0-9A-F]{2})*))?(?:#((?:[a-z0-9-._~!$&'()*+,;=:/?@]|%[0-9A-F]{2})*))?$`)

// ClientInfo レスポンス用クライアント情報構造体
type ClientInfo struct {
	ClientID    string `json:"clientId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatorID   string `json:"creatorId"`
}

// OwnedClientInfo レスポンス用クライアント情報構造体
type OwnedClientInfo struct {
	ClientID    string             `json:"clientId"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	CreatorID   string             `json:"creatorId"`
	Scopes      scope.AccessScopes `json:"scopes"`
	RedirectURI string             `json:"redirectUri"`
	Secret      string             `json:"secret"`
}

// AllowedClientInfo レスポンス用クライアント情報構造体
type AllowedClientInfo struct {
	TokenID     string             `json:"tokenId"`
	ClientID    string             `json:"clientId"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	CreatorID   string             `json:"creatorId"`
	Scopes      scope.AccessScopes `json:"scopes"`
	ApprovedAt  time.Time          `json:"approvedAt"`
}

// OAuth2APIHandler OAuth2のストアにアクセスするハンドラの集合
type OAuth2APIHandler struct {
	oauth2.Store
}

// GetMyTokens : GET /users/me/tokens
func (h *OAuth2APIHandler) GetMyTokens(c echo.Context) error {
	userID := c.Get("user").(*model.User).ID

	ot, err := h.Store.GetTokensByUser(uuid.FromStringOrNil(userID))
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	res := make([]*AllowedClientInfo, len(ot))
	for i, v := range ot {
		oc, err := h.Store.GetClient(v.ClientID)
		if err != nil {
			switch err {
			case oauth2.ErrClientNotFound:
				continue
			default:
				c.Logger().Error(err)
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
		}
		res[i] = &AllowedClientInfo{
			TokenID:     v.ID.String(),
			ClientID:    v.ClientID,
			Name:        oc.Name,
			Description: oc.Description,
			CreatorID:   oc.CreatorID.String(),
			Scopes:      v.Scopes,
			ApprovedAt:  v.CreatedAt,
		}
	}

	return c.JSON(http.StatusOK, res)
}

// DeleteMyToken : DELETE /users/me/tokens/:tokenID
func (h *OAuth2APIHandler) DeleteMyToken(c echo.Context) error {
	tokenID := c.Param("tokenID")
	userID := c.Get("user").(*model.User).ID

	ot, err := h.Store.GetTokenByID(uuid.FromStringOrNil(tokenID))
	if err != nil {
		switch err {
		case oauth2.ErrTokenNotFound:
			return echo.NewHTTPError(http.StatusNotFound)
		default:
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}

	if ot.UserID.String() != userID {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	if err := h.Store.DeleteTokenByAccess(ot.AccessToken); err != nil {
		switch err {
		case oauth2.ErrTokenNotFound:
			return echo.NewHTTPError(http.StatusNotFound)
		default:
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}

	return c.NoContent(http.StatusNoContent)
}

// GetClients : GET /clients
func (h *OAuth2APIHandler) GetClients(c echo.Context) error {
	userID := c.Get("user").(*model.User).ID

	oc, err := h.Store.GetClientsByUser(uuid.FromStringOrNil(userID))
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	res := make([]*OwnedClientInfo, len(oc))
	for i, v := range oc {
		res[i] = &OwnedClientInfo{
			ClientID:    v.ID,
			Name:        v.Name,
			Description: v.Description,
			CreatorID:   v.CreatorID.String(),
			Scopes:      v.Scopes,
			RedirectURI: v.RedirectURI,
			Secret:      v.Secret,
		}
	}

	return c.JSON(http.StatusOK, res)
}

// PostClients : POST /clients
func (h *OAuth2APIHandler) PostClients(c echo.Context) error {
	userID := c.Get("user").(*model.User).ID

	req := struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		RedirectURI string   `json:"redirectUri"`
		Scopes      []string `json:"scopes"`
	}{}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	if len(req.Name) == 0 || len(req.Name) > 32 {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	if len(req.Description) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	if !uriRegex.MatchString(req.RedirectURI) {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	scopes := scope.AccessScopes{}
	for _, v := range req.Scopes {
		s := scope.AccessScope(v)
		if !scope.Valid(s) {
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		scopes = append(scopes, s)
	}

	client := &oauth2.Client{
		ID:           uuid.NewV4().String(),
		Name:         req.Name,
		Description:  req.Description,
		Confidential: false,
		CreatorID:    uuid.FromStringOrNil(userID),
		RedirectURI:  req.RedirectURI,
		Secret:       base64.RawURLEncoding.EncodeToString(uuid.NewV4().Bytes()),
		Scopes:       scopes,
	}
	if err := h.Store.SaveClient(client); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, &OwnedClientInfo{
		ClientID:    client.ID,
		Name:        client.Name,
		Description: client.Description,
		CreatorID:   client.CreatorID.String(),
		Scopes:      client.Scopes,
		RedirectURI: client.RedirectURI,
		Secret:      client.Secret,
	})
}

// GetClient : GET /clients/:clientID
func (h *OAuth2APIHandler) GetClient(c echo.Context) error {
	clientID := c.Param("clientID")

	oc, err := h.Store.GetClient(clientID)
	if err != nil {
		switch err {
		case oauth2.ErrClientNotFound:
			return echo.NewHTTPError(http.StatusNotFound)
		default:
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}

	return c.JSON(http.StatusOK, &ClientInfo{
		ClientID:    oc.ID,
		Name:        oc.Name,
		Description: oc.Description,
		CreatorID:   oc.CreatorID.String(),
	})
}

// PatchClient : PATCH /clients/:clientID
func (h *OAuth2APIHandler) PatchClient(c echo.Context) error {
	clientID := c.Param("clientID")
	userID := c.Get("user").(*model.User).ID

	oc, err := h.Store.GetClient(clientID)
	if err != nil {
		switch err {
		case oauth2.ErrClientNotFound:
			return echo.NewHTTPError(http.StatusNotFound)
		default:
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}

	if oc.CreatorID.String() != userID {
		return echo.NewHTTPError(http.StatusForbidden)
	}

	req := struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		RedirectURI string `json:"redirectUri"`
	}{}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	if len(req.Name) > 0 {
		if len(req.Name) > 32 {
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		oc.Name = req.Name
	}

	if len(req.Description) > 0 {
		oc.Description = req.Description
	}

	if len(req.RedirectURI) > 0 {
		if !uriRegex.MatchString(req.RedirectURI) {
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		oc.RedirectURI = req.RedirectURI
	}

	if err := h.Store.UpdateClient(oc); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

// DeleteClient : DELETE /clients/:clientID
func (h *OAuth2APIHandler) DeleteClient(c echo.Context) error {
	clientID := c.Param("clientID")
	userID := c.Get("user").(*model.User).ID

	oc, err := h.Store.GetClient(clientID)
	if err != nil {
		switch err {
		case oauth2.ErrClientNotFound:
			return echo.NewHTTPError(http.StatusNotFound)
		default:
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}

	if oc.CreatorID.String() != userID {
		return echo.NewHTTPError(http.StatusForbidden)
	}

	// revoke tokens
	if err := h.Store.DeleteTokenByClient(clientID); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	// delete client
	if err := h.Store.DeleteClient(clientID); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}