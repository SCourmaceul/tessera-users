package dynamo

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/SCourmaceul/tetra-kit/adapter/dynamo"

	"github.com/SCourmaceul/tessera-users/internal/app"
	"github.com/SCourmaceul/tessera-users/internal/domain"
)

// fakeAPI est une table DynamoDB en mémoire (sous-ensemble utilisé par dynamo.Table).
type fakeAPI struct {
	items map[string]map[string]types.AttributeValue
}

func str(item map[string]types.AttributeValue, name string) string {
	if v, ok := item[name].(*types.AttributeValueMemberS); ok {
		return v.Value
	}
	return ""
}

func id(key map[string]types.AttributeValue) string {
	return str(key, "pk") + "|" + str(key, "sk")
}

func (f *fakeAPI) GetItem(_ context.Context, in *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return &dynamodb.GetItemOutput{Item: f.items[id(in.Key)]}, nil
}

func (f *fakeAPI) PutItem(_ context.Context, in *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	k := id(in.Item)
	if _, exists := f.items[k]; exists && in.ConditionExpression != nil {
		return nil, &types.ConditionalCheckFailedException{}
	}
	f.items[k] = in.Item
	return &dynamodb.PutItemOutput{}, nil
}

func (f *fakeAPI) DeleteItem(_ context.Context, in *dynamodb.DeleteItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	delete(f.items, id(in.Key))
	return &dynamodb.DeleteItemOutput{}, nil
}

func (f *fakeAPI) Query(_ context.Context, in *dynamodb.QueryInput, _ ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	pk, sk := str(in.ExpressionAttributeValues, ":pk"), str(in.ExpressionAttributeValues, ":sk")
	out := &dynamodb.QueryOutput{}
	for _, item := range f.items {
		if str(item, "pk") == pk && strings.HasPrefix(str(item, "sk"), sk) {
			out.Items = append(out.Items, item)
		}
	}
	return out, nil
}

func newStore() (Store, *fakeAPI) {
	api := &fakeAPI{items: map[string]map[string]types.AttributeValue{}}
	return Store{Table: dynamo.NewTable(api, "tessera-users")}, api
}

func TestProfileRoundTrip(t *testing.T) {
	ctx := context.Background()
	s, api := newStore()
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	p := domain.Profile{ID: "u-1", Pseudo: "ada", Spaces: []string{"core", "flux"}, CreatedAt: now, UpdatedAt: now}

	if err := s.CreateProfile(ctx, p); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateProfile(ctx, p); !errors.Is(err, app.ErrProfileExists) {
		t.Errorf("doublon : err = %v", err)
	}
	if _, ok := api.items["USER#u-1|PROFILE"]; !ok {
		t.Errorf("clé du profil inattendue : %v", api.items)
	}
	got, err := s.GetProfile(ctx, "u-1")
	if err != nil || got.Pseudo != "ada" || len(got.Spaces) != 2 || !got.CreatedAt.Equal(now) {
		t.Errorf("GetProfile = %+v, %v", got, err)
	}
	if _, err := s.GetProfile(ctx, "ghost"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("profil absent : err = %v", err)
	}
}

func TestMembers(t *testing.T) {
	ctx := context.Background()
	s, api := newStore()
	for _, m := range []domain.Member{
		{UserID: "u-1", Space: "core", Pseudo: "ada", Role: domain.RoleMember},
		{UserID: "u-2", Space: "core", Pseudo: "bob", Role: domain.RoleMember},
		{UserID: "u-1", Space: "flux", Pseudo: "ada", Role: domain.RoleMember},
	} {
		if err := s.CreateMember(ctx, m); err != nil {
			t.Fatal(err)
		}
	}
	if _, ok := api.items["SPACE#core#MEMBERS|USER#u-1"]; !ok {
		t.Errorf("clé du membre inattendue : %v", api.items)
	}
	if err := s.CreateMember(ctx, domain.Member{UserID: "u-1", Space: "core"}); !errors.Is(err, domain.ErrAlreadyMember) {
		t.Errorf("doublon : err = %v", err)
	}
	m, err := s.GetMember(ctx, "core", "u-2")
	if err != nil || m.Pseudo != "bob" || m.Space != "core" {
		t.Errorf("GetMember = %+v, %v", m, err)
	}
	if _, err := s.GetMember(ctx, "flow", "u-1"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("hors espace : err = %v", err)
	}
	members, err := s.ListMembers(ctx, "core")
	if err != nil || len(members) != 2 {
		t.Errorf("ListMembers(core) = %+v, %v", members, err)
	}
}
