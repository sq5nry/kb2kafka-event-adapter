package test

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/go-openapi/runtime"
	httpTransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/killbill/kbcli/v2/kbclient"
	"github.com/killbill/kbcli/v2/kbclient/account"
	"github.com/killbill/kbcli/v2/kbclient/tenant"
	"github.com/killbill/kbcli/v2/kbmodel"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"strconv"
	"testing"
)

var kbConfig = &kbConfiguration{
	host:      "killbill",
	port:      8080,
	user:      "admin",
	password:  "password",
	apiKey:    "api_key_test" + uuid.NewString(),
	apiSecret: "api_secret",
	callback:  "http://kb2kafka:8082/callmeback",
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupSuite() {
	suite.kbClient = createKbClient()

	createdBy := "integration_test"
	comment := "integration_test_comment"
	reason := "integration_test"

	suite.kbClient.SetDefaults(kbclient.KillbillDefaults{
		CreatedBy: &createdBy,
		Comment:   &comment,
		Reason:    &reason,
	})

	createTenant(suite)
	registerForNotifications(suite)
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	removeTenant(suite)
	unregisterForNotifications(suite)
}

// tests
func (suite *IntegrationTestSuite) TestCreateAndModifyAccount() {
	//given
	const id = "b41e63ce-f3d8-40a5-91ca-70d5831d4bf2"
	var accToCreate = &kbmodel.Account{
		AccountBalance:    1000,
		AccountCBA:        1,
		AccountID:         id,
		Address1:          "test_address",
		BillCycleDayLocal: 1,
		City:              "test_city",
		Company:           "test_company",
		Country:           "test_country",
		Currency:          "EUR",
		Email:             "test@test.com",
		Name:              id,
		Phone:             "+1234567890",
		PostalCode:        "00000",
		ReferenceTime:     strfmt.DateTime{},
	}
	newAccountReq := account.NewCreateAccountParams().WithBody(accToCreate)

	//when
	fmt.Printf("creating account, request=%#v\n", newAccountReq.Body)
	createResult, err := suite.kbClient.Account.CreateAccount(context.Background(), newAccountReq)
	if err != nil {
		panic(err)
	}
	fmt.Printf("create account, result=%d, data=%#v\n", createResult.HttpResponse.Code(), createResult.Payload)

	//then
	//TODO verify e2e flow: notification from a dedicated topic has arrived - these events are persisted in db:
	/* "TENANT_CONFIG_CHANGE", ObjectType:"TENANT_KVS", AccountId:"", 									  ObjectId:"40d0f0c7-c31a-4a9f-a1d8-63c03ec25071", MetaData:"CATALOG"}
       "ACCOUNT_CREATION",     ObjectType:"ACCOUNT",    AccountId:"d3ed72e0-b68a-4523-b4a4-3210cb2a6910", ObjectId:"d3ed72e0-b68a-4523-b4a4-3210cb2a6910", MetaData:""}
    */
}

// utilities
type IntegrationTestSuite struct {
	suite.Suite
	kbTestServer *testcontainers.LocalDockerCompose
	kbClient     *kbclient.KillBill
}

type kbConfiguration struct {
	host      string
	port      int
	user      string
	password  string
	apiKey    string
	apiSecret string
	callback  string
}

func createKbClient() *kbclient.KillBill {
	trp := httpTransport.New(kbConfig.host+":"+strconv.Itoa(kbConfig.port), "", nil)
	trp.Producers["text/xml"] = runtime.TextProducer()
	trp.Debug = true

	authWriter := runtime.ClientAuthInfoWriterFunc(func(r runtime.ClientRequest, _ strfmt.Registry) error {
		encoded := base64.StdEncoding.EncodeToString([]byte(kbConfig.user + ":" + kbConfig.password))
		if err := r.SetHeaderParam("Authorization", "Basic "+encoded); err != nil {
			return err
		}
		if err := r.SetHeaderParam("X-KillBill-ApiKey", kbConfig.apiKey); err != nil {
			return err
		}
		if err := r.SetHeaderParam("X-KillBill-ApiSecret", kbConfig.apiSecret); err != nil {
			return err
		}
		return nil
	})
	return kbclient.New(trp, strfmt.Default, authWriter, kbclient.KillbillDefaults{})
}

func registerForNotifications(suite *IntegrationTestSuite) {
	registerNotificationReq := tenant.NewRegisterPushNotificationCallbackParams().WithCb(&kbConfig.callback)
	callback, err := suite.kbClient.Tenant.RegisterPushNotificationCallback(context.Background(), registerNotificationReq)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("callback registered, result=%d\n", callback.HttpResponse.Code())
}

func unregisterForNotifications(suite *IntegrationTestSuite) {
	unregisterNotificationReq := tenant.NewDeletePushNotificationCallbacksParams()
	callback, err := suite.kbClient.Tenant.DeletePushNotificationCallbacks(context.Background(), unregisterNotificationReq)
	if err != nil {
		panic(err)
	}
	fmt.Printf("callback unregistered, result=%d\n", callback.HttpResponse.Code())
}

func createTenant(suite *IntegrationTestSuite) {
	tenant_id, _ := uuid.NewUUID()

	newTenantReq := tenant.NewCreateTenantParams().WithBody(&kbmodel.Tenant{
		APIKey:      &kbConfig.apiKey,
		APISecret:   &kbConfig.apiSecret,
		ExternalKey: tenant_id.String(),
	})

	fmt.Printf("creating tenant, request=%#v\n", newTenantReq.Body)
	createResult, err := suite.kbClient.Tenant.CreateTenant(context.Background(), newTenantReq)
	if err != nil {
		panic(err)
	}
	fmt.Printf("create tenant, result=%d\n", createResult.HttpResponse.Code())
}

func removeTenant(suite *IntegrationTestSuite) {
	deleteTenantReq := tenant.NewDeletePerTenantConfigurationParams()
	deleteResult, err := suite.kbClient.Tenant.DeletePerTenantConfiguration(context.Background(), deleteTenantReq)
	if err != nil {
		panic(err)
	}
	fmt.Printf("remove tenant, result=%d\n", deleteResult.HttpResponse.Code())
}
