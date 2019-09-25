package service

import (
	"runtime"
	"sync"
	"testing"
)

type MockLCASolver struct {
	raceMap map[int]int
}

func (solver *MockLCASolver) Setup([][]int) error {
	solver.raceMap = make(map[int]int)
	return nil
}

func (solver *MockLCASolver) SolveLCA(first, second int) (int, error) {
	for i := 0; i < 100; i++ {
		solver.raceMap[i] = i
		solver.raceMap[i] = solver.raceMap[i]
	}

	return first, nil
}

type corporateDirectoryTestCase struct {
	Left   int
	Right  int
	Answer *Employee
	Error  error
}

func validateCorporateDirectory(t *testing.T, dir CorporateDirectory, tests []corporateDirectoryTestCase) {
	for _, test := range tests {
		if ans, err := dir.GetCommonManager(test.Left, test.Right); ans != test.Answer || err != test.Error {
			errorFmt := "Test failed on test=(%d, %d); expected=%d; result=%d; error=%v; expected_error=%v"
			t.Errorf(errorFmt, test.Left, test.Right, test.Answer, ans, err, test.Error)
		}
	}
}

func TestCorporateDirectoryServiceSetup(t *testing.T) {
	employees := []*Employee{
		{1, "Claire", []int{}},
	}
	dir := NewCorporateDirectoryService(&MockLCASolver{})

	err := dir.Setup(employees)
	if err != nil {
		t.Error("setup failed")
	}
}

func TestCorporateDirectoryServiceSetupDuplicatedId(t *testing.T) {
	employees := []*Employee{
		{1, "Claire", []int{}},
		{1, "A", []int{}},
	}
	dir := NewCorporateDirectoryService(&MockLCASolver{})

	err := dir.Setup(employees)
	if err != ErrEmployeeExists {
		t.Error("setup failed")
	}
}

func TestCorporateDirectoryServiceSetupInvalidEdge(t *testing.T) {
	employees := []*Employee{
		{2, "A", []int{5}},
		{1, "Claire", []int{2}},
		{3, "B", []int{}},
	}
	dir := NewCorporateDirectoryService(&MockLCASolver{})

	err := dir.Setup(employees)
	if err != ErrInvalidEdge {
		t.Error("setup failed")
	}
}

func TestCorporateDirectoryServiceSetupBossNotFound(t *testing.T) {
	employees := []*Employee{
		{1, "_", []int{}},
		{2, "A", []int{}},
		{3, "B", []int{}},
	}
	dir := NewCorporateDirectoryService(&MockLCASolver{})

	err := dir.Setup(employees)
	if err != ErrBossNotFound {
		t.Error("setup failed")
	}
}

func TestCorporateDirectoryServiceBasicLookup(t *testing.T) {
	employees := []*Employee{
		{1, "Claire", []int{1, 2}},
		{2, "A", []int{}},
		{3, "B", []int{}},
	}

	tests := []corporateDirectoryTestCase{
		{1, 3, employees[0], nil},
		{2, 3, employees[1], nil},
		{3, 3, employees[2], nil},
	}

	dir := NewCorporateDirectoryService(&MockLCASolver{})

	err := dir.Setup(employees)
	if err != nil {
		t.Error("setup failed")
	}
	validateCorporateDirectory(t, dir, tests)
}

func TestCorporateDirectoryServiceBasicLookupDifferentOrder(t *testing.T) {
	employees := []*Employee{
		{3, "B", []int{}},
		{1, "Claire", []int{1, 2}},
		{2, "A", []int{}},
	}

	tests := []corporateDirectoryTestCase{
		{1, 3, employees[1], nil},
		{2, 3, employees[2], nil},
		{3, 3, employees[0], nil},
	}

	dir := NewCorporateDirectoryService(&MockLCASolver{})

	err := dir.Setup(employees)
	if err != nil {
		t.Error("setup failed")
	}
	validateCorporateDirectory(t, dir, tests)
}

func TestCorporateDirectoryServiceInvalidId(t *testing.T) {
	employees := []*Employee{
		{3, "B", []int{}},
		{1, "Claire", []int{1, 2}},
		{2, "A", []int{}},
	}

	tests := []corporateDirectoryTestCase{
		{0, 3, nil, ErrInvalidEmployee},
		{2, 0, nil, ErrInvalidEmployee},
	}

	dir := NewCorporateDirectoryService(&MockLCASolver{})

	err := dir.Setup(employees)
	if err != nil {
		t.Error("setup failed")
	}
	validateCorporateDirectory(t, dir, tests)
}

func TestCorporateDirectoryServiceSetupRaceCondition(t *testing.T) {
	wg := &sync.WaitGroup{}
	dir := NewCorporateDirectoryService(&MockLCASolver{})

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 100000; j++ {
				employees := []*Employee{
					{1, "B", []int{}},
					{2, "Claire", []int{1, 3, 4, 5, 6}},
					{3, "A", []int{}},
					{4, "A", []int{}},
					{5, "A", []int{}},
					{6, "A", []int{}},
				}

				err := dir.Setup(employees)
				if err != nil {
					t.Error("setup failed")
				}
				runtime.Gosched()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestCorporateDirectoryServiceRWRaceCondition(t *testing.T) {
	wg := &sync.WaitGroup{}
	dir := NewCorporateDirectoryService(&MockLCASolver{})

	for i := 0; i < 8; i++ {
		wg.Add(2)
		go func() {
			for j := 0; j < 100000; j++ {
				employees := []*Employee{
					{1, "B", []int{}},
					{2, "Claire", []int{1, 3, 4, 5, 6}},
					{3, "A", []int{}},
					{4, "A", []int{}},
					{5, "A", []int{}},
					{6, "A", []int{}},
				}

				err := dir.Setup(employees)
				if err != nil {
					t.Error("setup failed")
				}
				runtime.Gosched()
			}
			wg.Done()
		}()

		go func() {
			for j := 0; j < 100000; j++ {
				_, _ = dir.GetCommonManager(0, 0)
				runtime.Gosched()
			}
			wg.Done()
		}()

	}
	wg.Wait()
}
