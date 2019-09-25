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

type CorporateDirectory interface {
	Setup(employees []*Employee) error
	GetCommonManager(first, second int) (*Employee, error)
	GetEmployee(id int) (*Employee, error)
	GetEmployees() ([]*Employee, error)
}

type CorporateDirectoryService struct {
	idToIndex *sync.Map

	employees []*Employee

	setupMutex sync.RWMutex
	solver     lca.LCASolver
}

func NewCorporateDirectoryService(solver lca.LCASolver) *CorporateDirectoryService {
	return &CorporateDirectoryService{
		idToIndex: &sync.Map{},
		solver:    solver,
	}
}

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

	idToIndex := sync.Map{}
	for idx, employee := range employees {
		if _, ok := idToIndex.Load(employee.ID); ok {
			return ErrEmployeeExists
		}
		idToIndex.Store(employee.ID, idx)
	}

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

	err := dir.solver.Setup(nodesAdjList)
	if err != nil {
		return err
	}

	dir.idToIndex = &idToIndex
	dir.employees = employees
	return nil
}

func (dir *CorporateDirectoryService) GetCommonManager(first, second int) (*Employee, error) {
	dir.setupMutex.RLock()
	defer dir.setupMutex.RUnlock()

	firstId, err := dir.resolveId(first)
	if err != nil {
		return nil, err
	}
	secondId, err := dir.resolveId(second)
	if err != nil {
		return nil, err
	}

	commonId, err := dir.solver.SolveLCA(firstId, secondId)
	if err != nil {
		return nil, err
	}

	return dir.employees[commonId], nil
}

func (dir *CorporateDirectoryService) GetEmployee(id int) (*Employee, error) {
	dir.setupMutex.RLock()
	defer dir.setupMutex.RUnlock()

	employeeId, err := dir.resolveId(id)
	if err != nil {
		return nil, err
	}

	return dir.employees[employeeId], nil
}

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
