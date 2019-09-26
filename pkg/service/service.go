package service

import (
	"corporate-directory/pkg/lca"
	"errors"
	"sync"
)

var (
	ErrInvalidEmployee = errors.New(`employee with given id was not found`)
	ErrInvalidEdge     = errors.New(`employee links to an invalid employee id`)
	ErrEmployeeExists  = errors.New(`multiple employees with same id`)
	ErrBossNotFound    = errors.New(`employee with name Claire was not found`)
)

type Employee struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Subordinates []int  `json:"subordinates"`
}

// We assume that employees are known in advance or change rarely so we can afford to recalculate the solution
// For tests we will be able to mock the service or swap the implementation
type CorporateDirectory interface {
	Setup(employees []*Employee) error
	GetCommonManager(first, second int) (*Employee, error)
	GetEmployee(id int) (*Employee, error)
	GetEmployees() ([]*Employee, error)
}

// Service implementation. Main functionality implemented by this service is ID resolution from client representation
// to representation required by LCASolver interface
type CorporateDirectoryService struct {
	// Map to lookup employee index by his/her ID
	idToIndex *sync.Map

	// employees list
	employees []*Employee

	// Lock so we don't get into race conditions with simultaneous setup/common requests
	setupMutex sync.RWMutex
	// Solver implementation injected into this service
	solver lca.LCASolver
}

func NewCorporateDirectoryService(solver lca.LCASolver) *CorporateDirectoryService {
	return &CorporateDirectoryService{
		idToIndex: &sync.Map{},
		solver:    solver,
	}
}

// Setup service, preparing data structures for further queries
func (dir *CorporateDirectoryService) Setup(employees []*Employee) error {
	dir.setupMutex.Lock()
	defer dir.setupMutex.Unlock()

	// Find Claire and place her as the first node
	ok := false
	for idx, employee := range employees {
		if employee.Name == "Claire" {
			employees[0], employees[idx] = employees[idx], employees[0]
			ok = true
			break
		}
	}
	if !ok {
		return ErrBossNotFound
	}

	// Prepare Employee ID -> Employee index in array map, making sure Employee IDs are unique
	idToIndex := sync.Map{}
	for idx, employee := range employees {
		if _, ok := idToIndex.Load(employee.ID); ok {
			return ErrEmployeeExists
		}
		idToIndex.Store(employee.ID, idx)
	}

	// Prepare represenstion for LCASolver interface  while checking all the edges
	nodesAdjList := make([][]int, len(employees))
	for idx, node := range employees {
		for _, child := range node.Subordinates {
			childNodeId, ok := idToIndex.Load(child)
			if !ok {
				return ErrInvalidEdge
			}
			nodesAdjList[idx] = append(nodesAdjList[idx], childNodeId.(int))
		}
	}

	// Setup solver and if everything went well update service struct
	err := dir.solver.Setup(nodesAdjList)
	if err != nil {
		return err
	}

	dir.idToIndex = &idToIndex
	dir.employees = employees
	return nil
}

// Actual request, get closest common manager for two employees by their ID
func (dir *CorporateDirectoryService) GetCommonManager(first, second int) (*Employee, error) {
	dir.setupMutex.RLock()
	defer dir.setupMutex.RUnlock()

	// Resolve indices
	firstId, err := dir.resolveId(first)
	if err != nil {
		return nil, err
	}
	secondId, err := dir.resolveId(second)
	if err != nil {
		return nil, err
	}

	// Find solution and return corresponding employee
	commonId, err := dir.solver.SolveLCA(firstId, secondId)
	if err != nil {
		return nil, err
	}

	return dir.employees[commonId], nil
}

// Convenience method to get an employee by ID
func (dir *CorporateDirectoryService) GetEmployee(id int) (*Employee, error) {
	dir.setupMutex.RLock()
	defer dir.setupMutex.RUnlock()

	employeeId, err := dir.resolveId(id)
	if err != nil {
		return nil, err
	}

	return dir.employees[employeeId], nil
}

// Method to list all employees registered in the system
func (dir *CorporateDirectoryService) GetEmployees() ([]*Employee, error) {
	dir.setupMutex.RLock()
	defer dir.setupMutex.RUnlock()
	return dir.employees, nil
}

func (dir *CorporateDirectoryService) resolveId(first int) (int, error) {
	id, ok := dir.idToIndex.Load(first)
	if !ok {
		return 0, ErrInvalidEmployee
	}
	return id.(int), nil
}
