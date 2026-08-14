//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/generated/recapapi"
	recapcontroller "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/controller/http/recap"
	activityrepo "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/repository/activity"
	listingrepo "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/repository/listing"
	profilerepo "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/repository/profile"
	recaprepo "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/repository/recap"
	sharedrecaprepo "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/repository/sharedrecap"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/seed"
	profileusecase "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/usecase/profile"
	recapusecase "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/usecase/recap"
	sharedrecapusecase "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/usecase/sharedrecap"
)

const (
	recapYear = 2026
	seedValue = 20260807
)

func TestRecapJourney(t *testing.T) {
	generateDemoData(t)
	server := newAPIServer(t)
	client := server.Client()

	profileID := findProfile(t, client, server.URL, "Анна", "Воронова")
	recapID := createRecap(t, client, server.URL, profileID, http.StatusCreated)
	assertPersonalRecap(t, client, server.URL, recapID)

	existingRecapID := createRecap(t, client, server.URL, profileID, http.StatusOK)
	assert.Equal(t, recapID, existingRecapID)

	shared := shareRecap(t, client, server.URL, recapID, http.StatusCreated)
	assertPublicRecap(t, client, server.URL, shared.Token)

	existingShared := shareRecap(t, client, server.URL, recapID, http.StatusOK)
	assert.Equal(t, shared.Token, existingShared.Token)
	assert.Equal(t, shared.URL.String(), existingShared.URL.String())
}

func generateDemoData(t *testing.T) {
	t.Helper()

	generator, err := seed.NewGenerator(recapYear, seedValue, seed.DefaultCatalog())
	require.NoError(t, err)

	dataset, err := generator.Generate(seed.DefaultScenarios())
	require.NoError(t, err)
	require.NoError(t, seed.NewWriter(testEnv.pool).Write(testContext(t), dataset, false))
}

func newAPIServer(t *testing.T) *httptest.Server {
	t.Helper()

	activityRepository := activityrepo.New(testEnv.pool, operationTimeout)
	listingRepository := listingrepo.New(testEnv.pool, operationTimeout)
	profileRepository := profilerepo.New(testEnv.pool, operationTimeout)
	recapRepository := recaprepo.New(testEnv.pool, operationTimeout)
	sharedRepository := sharedrecaprepo.New(testEnv.pool, operationTimeout)

	recapService := recapusecase.NewRecapService(
		activityRepository,
		recapRepository,
		profileRepository,
		listingRepository,
	)
	profileService := profileusecase.NewProfileService(profileRepository)
	sharedService := sharedrecapusecase.NewSharedRecapService(
		recapRepository,
		profileRepository,
		sharedRepository,
		sharedrecapusecase.GenerateToken,
	)

	handler := recapcontroller.NewRecapServer(
		zap.NewNop(),
		recapService,
		profileService,
		sharedService,
		url.URL{Scheme: "https", Host: "recap50.ru"},
	)
	apiServer, err := recapapi.NewServer(handler)
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiServer))
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server
}

func findProfile(
	t *testing.T,
	client *http.Client,
	baseURL string,
	name string,
	surname string,
) uuid.UUID {
	t.Helper()

	response := doRequest(t, client, http.MethodGet, baseURL+"/api/v1/profiles", nil)
	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)

	var body struct {
		Items []recapapi.Profile `json:"items"`
	}
	decodeJSON(t, response.Body, &body)
	require.Len(t, body.Items, len(seed.DefaultScenarios()))

	for _, profile := range body.Items {
		if profile.Name == name && profile.Surname == surname {
			return uuid.UUID(profile.ID)
		}
	}

	require.FailNow(t, "demo profile not found", "%s %s", name, surname)

	return uuid.Nil
}

func createRecap(
	t *testing.T,
	client *http.Client,
	baseURL string,
	profileID uuid.UUID,
	wantStatus int,
) uuid.UUID {
	t.Helper()

	payload := map[string]any{"profileId": profileID, "year": recapYear}
	response := doJSONRequest(t, client, http.MethodPost, baseURL+"/api/v1/recaps", payload)
	defer response.Body.Close()
	require.Equal(t, wantStatus, response.StatusCode)

	var body recapapi.CreateRecapResponse
	decodeJSON(t, response.Body, &body)

	return uuid.UUID(body.ID)
}

func assertPersonalRecap(t *testing.T, client *http.Client, baseURL string, recapID uuid.UUID) {
	t.Helper()

	response := doRequest(
		t,
		client,
		http.MethodGet,
		baseURL+"/api/v1/recaps/"+recapID.String(),
		nil,
	)
	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)

	var body recapapi.Recap
	decodeJSON(t, response.Body, &body)
	assert.Equal(t, recapapi.UUID(recapID), body.ID)
	assert.Equal(t, recapapi.RecapStatusReady, body.Status)
	assert.Equal(t, recapapi.ArchetypeCodeCollector, body.Archetype.Code)
	assert.NotEmpty(t, body.Archetype.Reasons)
	assert.NotEmpty(t, body.Slides)
}

func shareRecap(
	t *testing.T,
	client *http.Client,
	baseURL string,
	recapID uuid.UUID,
	wantStatus int,
) recapapi.SharedRecapLink {
	t.Helper()

	response := doRequest(
		t,
		client,
		http.MethodPost,
		baseURL+"/api/v1/recaps/"+recapID.String()+"/share",
		nil,
	)
	defer response.Body.Close()
	require.Equal(t, wantStatus, response.StatusCode)

	var body recapapi.SharedRecapLink
	decodeJSON(t, response.Body, &body)
	assert.NotEmpty(t, body.Token)
	assert.Equal(t, "https://recap50.ru/shared-recaps/"+string(body.Token), body.URL.String())

	return body
}

func assertPublicRecap(
	t *testing.T,
	client *http.Client,
	baseURL string,
	token recapapi.SharedRecapToken,
) {
	t.Helper()

	response := doRequest(
		t,
		client,
		http.MethodGet,
		baseURL+"/api/v1/shared-recaps/"+string(token),
		nil,
	)
	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)

	raw, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "Воронова")

	var fields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &fields))
	for _, privateField := range []string{
		"profileId",
		"recapId",
		"reasons",
		"favorites",
		"purchases",
		"sales",
		"messages",
		"recommendations",
		"amountRange",
	} {
		assert.NotContains(t, fields, privateField)
	}

	var body recapapi.SharedRecap
	require.NoError(t, json.Unmarshal(raw, &body))
	assert.Equal(t, "Анна", body.DisplayName)
	assert.Equal(t, recapapi.SharedArchetypeCodeCollector, body.Archetype.Code)
	assert.Positive(t, body.ActiveDays)
}

func doJSONRequest(
	t *testing.T,
	client *http.Client,
	method string,
	target string,
	body any,
) *http.Response {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	return doRequest(t, client, method, target, bytes.NewReader(payload))
}

func doRequest(
	t *testing.T,
	client *http.Client,
	method string,
	target string,
	body io.Reader,
) *http.Response {
	t.Helper()

	request, err := http.NewRequestWithContext(t.Context(), method, target, body)
	require.NoError(t, err)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := client.Do(request)
	require.NoError(t, err)

	return response
}

func decodeJSON(t *testing.T, reader io.Reader, target any) {
	t.Helper()

	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	require.NoError(t, decoder.Decode(target))
}
