package di_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rom8726/di"
)

type DBClient interface {
	Exec() (string, error)
}

type DBClientImpl struct {
	data string
}

func (c *DBClientImpl) Exec() (string, error) { return c.data, nil }

func NewDBClient(data string) *DBClientImpl {
	return &DBClientImpl{data: data}
}

type Repo interface {
	Find() (string, error)
}

type RepoImpl struct {
	db DBClient
}

func (r *RepoImpl) Find() (string, error) { return r.db.Exec() }

func NewRepo(db DBClient) *RepoImpl {
	return &RepoImpl{db: db}
}

type Service interface {
	Run() (string, error)
}

type MyService struct {
	params *MyServiceParams
	repo   Repo
}

type MyServiceParams struct {
	ParamInt  int
	ParamStr  string
	ParamBool bool
}

func NewMyService(params *MyServiceParams, r Repo) *MyService {
	return &MyService{params: params, repo: r}
}

func (s *MyService) Run() (string, error) {
	data, err := s.repo.Find()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Running MyService(%+v) with: %s", s.params, data), nil
}

type Service2 interface {
	Run2() (string, error)
}

type MyService2 struct {
	repo Repo

	paramInt  int
	paramBool bool
}

func NewMyService2(paramInt int, paramBool bool, r Repo) *MyService2 {
	return &MyService2{repo: r, paramInt: paramInt, paramBool: paramBool}
}

func (s *MyService2) Run2() (string, error) {
	data, err := s.repo.Find()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Running MyService2(int: %d, bool: %v) with: %s", s.paramInt, s.paramBool, data), nil
}

type RootService struct {
	service  Service
	service2 Service2
}

func NewRootService(s Service, s2 Service2) *RootService {
	return &RootService{service: s, service2: s2}
}

func (s *RootService) RunServices() string {
	data1, _ := s.service.Run()
	data2, _ := s.service2.Run2()

	return fmt.Sprintf("Running RootService:\n%s\n%s", data1, data2)
}

type RootSrv interface {
	RunServices() string
}

// ---circular deps ---
type Dep1 interface {
	Do1()
}

type Dep2 interface {
	Do2()
}

type Dep1Impl struct {
	dep2 Dep2
}

func newDep1(dep2 Dep2) *Dep1Impl {
	return &Dep1Impl{dep2: dep2}
}

func (d *Dep1Impl) Do1() {
	d.dep2.Do2()
}

type Dep2Impl struct {
	dep1 Dep1
}

func newDep2(dep1 Dep1) *Dep2Impl {
	return &Dep2Impl{dep1: dep1}
}

func (d *Dep2Impl) Do2() {
	d.dep1.Do1()
}

func TestContainer_ResolveWithInterface(t *testing.T) {
	c := di.New()
	c.Provide(NewDBClient).Arg("data")
	c.Provide(NewRepo)
	c.Provide(NewMyService).Arg(&MyServiceParams{ParamInt: 1, ParamStr: "str", ParamBool: true})
	c.Provide(NewMyService2).Args(2, true)
	c.Provide(NewRootService)

	var rootSrv RootSrv
	err := c.Resolve(&rootSrv)
	require.NoError(t, err)

	actual := rootSrv.RunServices()
	expected := `Running RootService:
Running MyService(&{ParamInt:1 ParamStr:str ParamBool:true}) with: data
Running MyService2(int: 2, bool: true) with: data`
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestContainer_ResolveWithoutInterface(t *testing.T) {
	c := di.New()
	c.Provide(NewDBClient).Arg("data")
	c.Provide(NewRepo)
	c.Provide(NewMyService).Arg(&MyServiceParams{ParamInt: 1, ParamStr: "str", ParamBool: true})
	c.Provide(NewMyService2).Args(2, true)
	c.Provide(NewRootService)

	var rootSrv *RootService
	err := c.Resolve(&rootSrv)
	require.NoError(t, err)

	actual := rootSrv.RunServices()
	expected := `Running RootService:
Running MyService(&{ParamInt:1 ParamStr:str ParamBool:true}) with: data
Running MyService2(int: 2, bool: true) with: data`
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestContainer_ResolveWithCircularDep(t *testing.T) {
	c := di.New()
	c.Provide(newDep1)
	c.Provide(newDep2)

	var dep1 Dep1
	err := c.Resolve(&dep1)
	require.Error(t, err)
}

func TestContainer_ResolveToStruct(t *testing.T) {
	c := di.New()
	c.Provide(NewDBClient).Arg("data")
	c.Provide(NewRepo)
	c.Provide(NewMyService).Arg(&MyServiceParams{ParamInt: 1, ParamStr: "str", ParamBool: true})
	c.Provide(NewMyService2).Args(2, true)
	c.Provide(NewRootService)

	type Holder struct {
		DBClient *DBClientImpl
		Repo     *RepoImpl
		Service  *MyService
		Service2 *MyService2
		Root     *RootService
	}

	var holder Holder
	err := c.ResolveToStruct(&holder)
	require.NoError(t, err)

	require.NotNil(t, holder.DBClient)
	require.NotNil(t, holder.Repo)
	require.NotNil(t, holder.Service)
	require.NotNil(t, holder.Service2)
	require.NotNil(t, holder.Root)
}
