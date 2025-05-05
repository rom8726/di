package di_test

import (
	"fmt"
	"testing"

	"github.com/rom8726/di"
)

type DBClient interface {
	Exec() error
}

type DBClientImpl struct{}

func (c *DBClientImpl) Exec() error { return nil }

func NewDBClient() *DBClientImpl {
	return &DBClientImpl{}
}

type Repo interface {
	Find() string
}

type RepoImpl struct {
	db DBClient
}

func (r *RepoImpl) Find() string { return "data" }

func NewRepo(db DBClient) *RepoImpl {
	return &RepoImpl{db: db}
}

type Service interface {
	Run() string
}

type MyService struct {
	repo Repo
}

func NewService(r Repo) *MyService {
	return &MyService{repo: r}
}

func (s *MyService) Run() string {
	return fmt.Sprint("Running MyService with: ", s.repo.Find())
}

type Service2 interface {
	Run2() string
}

type MyService2 struct {
	repo Repo
}

func NewService2(r Repo) *MyService2 {
	return &MyService2{repo: r}
}

func (s *MyService2) Run2() string {
	return fmt.Sprint("Running MyService2 with: ", s.repo.Find())
}

type RootService struct {
	service  Service
	service2 Service2
}

func NewRootService(s Service, s2 Service2) *RootService {
	return &RootService{service: s, service2: s2}
}

func (s *RootService) RunServices() string {
	return fmt.Sprintf("Running RootService:\n%s\n%s", s.service.Run(), s.service2.Run2())
}

type RootSrv interface {
	RunServices() string
}

func TestContainer_ResolveWithInterface(t *testing.T) {
	c := di.New()
	c.Provide(NewDBClient)
	c.Provide(NewRepo)
	c.Provide(NewService)
	c.Provide(NewService2)
	c.Provide(NewRootService)

	var rootSrv RootSrv
	err := c.Resolve(&rootSrv)
	if err != nil {
		t.Fatal(err)
	}

	actual := rootSrv.RunServices()
	expected := `Running RootService:
Running MyService with: data
Running MyService2 with: data`
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestContainer_ResolveWithoutInterface(t *testing.T) {
	c := di.New()
	c.Provide(NewDBClient)
	c.Provide(NewRepo)
	c.Provide(NewService)
	c.Provide(NewService2)
	c.Provide(NewRootService)

	var rootSrv *RootService
	err := c.Resolve(&rootSrv)
	if err != nil {
		t.Fatal(err)
	}

	actual := rootSrv.RunServices()
	expected := `Running RootService:
Running MyService with: data
Running MyService2 with: data`
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}
