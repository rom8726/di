package di_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/rom8726/di"
)

type AppService1 interface {
	Do1()
}

type AppService2 interface {
	Do2()
}

type MockAppService1 struct {
	mock.Mock
}

func (m *MockAppService1) Do1() {
	m.Called()
}

func (m *MockAppService1) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAppService1) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Mock for AppService2
type MockAppService2 struct {
	mock.Mock
}

func (m *MockAppService2) Do2() {
	m.Called()
}

func (m *MockAppService2) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAppService2) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestApp_Start(t *testing.T) {
	errStart := errors.New("start error")

	tests := []struct {
		name         string
		startTimeout time.Duration
		contextTime  time.Duration
		expectedErr  error
		setupMocks   func(mock1 *MockAppService1, mock2 *MockAppService2)
	}{
		{
			name:         "success",
			startTimeout: 0,
			contextTime:  time.Second,
			expectedErr:  nil,
			setupMocks: func(mock1 *MockAppService1, mock2 *MockAppService2) {
				mock1.On("Start", mock.Anything).Return(nil)
				mock2.On("Start", mock.Anything).Return(nil)
			},
		},
		{
			name:         "start timeout",
			startTimeout: 100 * time.Millisecond,
			contextTime:  time.Second,
			expectedErr:  context.DeadlineExceeded,
			setupMocks: func(mock1 *MockAppService1, mock2 *MockAppService2) {
				mock1.On("Start", mock.Anything).After(time.Second).Return(nil)
				mock2.On("Start", mock.Anything).Return(nil)
			},
		},
		{
			name:         "start error",
			startTimeout: 0,
			contextTime:  time.Second,
			expectedErr:  errStart,
			setupMocks: func(mock1 *MockAppService1, mock2 *MockAppService2) {
				mock1.On("Start", mock.Anything).Return(errStart)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService1 := &MockAppService1{}
			mockService2 := &MockAppService2{}

			test.setupMocks(mockService1, mockService2)

			container := di.New()
			container.Provide(func() AppService1 { return mockService1 })
			container.Provide(func() AppService2 { return mockService2 })

			var srv1 AppService1
			if err := container.Resolve(&srv1); err != nil {
				t.Fatal(err)
			}

			var srv2 AppService2
			if err := container.Resolve(&srv2); err != nil {
				t.Fatal(err)
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			app := di.NewApp(container, di.WithLogger(logger), di.WithStartTimeout(test.startTimeout))

			ctx, cancel := context.WithTimeout(context.Background(), test.contextTime)
			defer cancel()

			// Start app and check errors
			err := app.Start(ctx)
			if test.expectedErr != nil {
				if !errors.Is(err, test.expectedErr) {
					t.Errorf("expected error %v, got %v", test.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			// Assert mock expectations
			mockService1.AssertExpectations(t)
			mockService2.AssertExpectations(t)
		})
	}
}

func TestApp_Stop(t *testing.T) {
	errStop := errors.New("stop error")
	tests := []struct {
		name        string
		stopTimeout time.Duration
		contextTime time.Duration
		expectedErr error
		setupMocks  func(mock1 *MockAppService1, mock2 *MockAppService2)
	}{
		{
			name:        "success",
			stopTimeout: 0,
			contextTime: time.Second,
			expectedErr: nil,
			setupMocks: func(mock1 *MockAppService1, mock2 *MockAppService2) {
				mock1.On("Stop", mock.Anything).Return(nil)
				mock2.On("Stop", mock.Anything).Return(nil)
			},
		},
		{
			name:        "stop timeout",
			stopTimeout: 100 * time.Millisecond,
			contextTime: time.Second,
			expectedErr: nil,
			setupMocks: func(mock1 *MockAppService1, mock2 *MockAppService2) {
				mock1.On("Stop", mock.Anything).After(time.Second).Return(nil)
				mock2.On("Stop", mock.Anything).Return(nil)
			},
		},
		{
			name:        "stop error",
			stopTimeout: 0,
			contextTime: time.Second,
			expectedErr: errStop,
			setupMocks: func(mock1 *MockAppService1, mock2 *MockAppService2) {
				mock1.On("Stop", mock.Anything).Return(errStop)
				mock2.On("Stop", mock.Anything).Return(nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService1 := &MockAppService1{}
			mockService2 := &MockAppService2{}

			test.setupMocks(mockService1, mockService2)

			container := di.New()
			container.Provide(func() AppService1 { return mockService1 })
			container.Provide(func() AppService2 { return mockService2 })

			var srv1 AppService1
			if err := container.Resolve(&srv1); err != nil {
				t.Fatal(err)
			}

			var srv2 AppService2
			if err := container.Resolve(&srv2); err != nil {
				t.Fatal(err)
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			app := di.NewApp(container, di.WithLogger(logger), di.WithStopTimeout(test.stopTimeout))

			ctx, cancel := context.WithTimeout(context.Background(), test.contextTime)
			defer cancel()

			// Stop app and check errors
			err := app.Stop(ctx)
			if test.expectedErr != nil {
				if !errors.Is(err, test.expectedErr) {
					t.Errorf("expected error %v, got %v", test.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			// Assert mock expectations
			mockService1.AssertExpectations(t)
			mockService2.AssertExpectations(t)
		})
	}
}

func TestApp_Run(t *testing.T) {
	errStart := errors.New("start error")
	//errStop := errors.New("stop error")

	tests := []struct {
		name         string
		startTimeout time.Duration
		stopTimeout  time.Duration
		contextTime  time.Duration
		expectedErr  error
		setupMocks   func(mock1 *MockAppService1, mock2 *MockAppService2)
	}{
		{
			name:         "run success",
			startTimeout: 0,
			stopTimeout:  time.Second,
			contextTime:  2 * time.Second,
			expectedErr:  nil,
			setupMocks: func(mock1 *MockAppService1, mock2 *MockAppService2) {
				mock1.On("Start", mock.Anything).Return(nil)
				mock1.On("Stop", mock.Anything).Return(nil)
				mock2.On("Start", mock.Anything).Return(nil)
				mock2.On("Stop", mock.Anything).Return(nil)
			},
		},
		{
			name:         "start timeout",
			startTimeout: 100 * time.Millisecond,
			stopTimeout:  time.Second,
			contextTime:  2 * time.Second,
			expectedErr:  context.DeadlineExceeded,
			setupMocks: func(mock1 *MockAppService1, mock2 *MockAppService2) {
				mock1.On("Start", mock.Anything).After(time.Second).Return(nil)
				mock1.On("Stop", mock.Anything).Return(nil)
				mock2.On("Stop", mock.Anything).Return(nil)
			},
		},
		{
			name:         "start error",
			startTimeout: time.Second,
			stopTimeout:  time.Second,
			contextTime:  2 * time.Second,
			expectedErr:  errStart,
			setupMocks: func(mock1 *MockAppService1, mock2 *MockAppService2) {
				mock1.On("Start", mock.Anything).Return(errStart)
				mock1.On("Stop", mock.Anything).Return(nil)
				mock2.On("Stop", mock.Anything).Return(nil)
			},
		},
		{
			name:         "context canceled during run",
			startTimeout: time.Second,
			stopTimeout:  time.Second,
			contextTime:  500 * time.Millisecond,
			expectedErr:  nil,
			setupMocks: func(mock1 *MockAppService1, mock2 *MockAppService2) {
				mock1.On("Start", mock.Anything).Return(nil)
				mock1.On("Stop", mock.Anything).Return(nil)
				mock2.On("Start", mock.Anything).Return(nil)
				mock2.On("Stop", mock.Anything).Return(nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService1 := &MockAppService1{}
			mockService2 := &MockAppService2{}

			test.setupMocks(mockService1, mockService2)

			container := di.New()
			container.Provide(func() AppService1 { return mockService1 })
			container.Provide(func() AppService2 { return mockService2 })

			var srv1 AppService1
			if err := container.Resolve(&srv1); err != nil {
				t.Fatal(err)
			}

			var srv2 AppService2
			if err := container.Resolve(&srv2); err != nil {
				t.Fatal(err)
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			app := di.NewApp(container, di.WithLogger(logger), di.WithStartTimeout(test.startTimeout), di.WithStopTimeout(test.stopTimeout))

			ctx, cancel := context.WithTimeout(context.Background(), test.contextTime)
			defer cancel()

			// Run the app and check errors
			err := app.Run(ctx)
			if test.expectedErr != nil {
				if !errors.Is(err, test.expectedErr) {
					t.Errorf("expected error %v, got %v", test.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			// Assert mock expectations
			mockService1.AssertExpectations(t)
			mockService2.AssertExpectations(t)
		})
	}
}
