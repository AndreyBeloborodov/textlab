package training

import (
    "errors"

    "github.com/AndreyBeloborodov/textlab/algorithm"

)

var (
    ErrTrainingInProgress   = errors.New("training is already in progress")
    ErrAlgorithmUnavailable = errors.New("algorithm factory is not configured")
)

type Manager interface {
    Submit(trainingID int) error
}

type submitRequest struct {
    trainingID int
    result     chan error
}

type trainingResult struct {
    trainingID int
    state      []byte
    err        error
}

type manager struct {
    repo    Repository
    factory algorithm.Factory
    store   algorithm.Store
    
    submitCh chan submitRequest
    resultCh chan trainingResult
}

func NewManager(
    repo Repository,
    factory algorithm.Factory,
    store algorithm.Store,
    ) Manager {
    m := &manager{
        repo:     repo,
        factory:  factory,
        store:    store,
        submitCh: make(chan submitRequest),
        resultCh: make(chan trainingResult),
    }

    go m.run()
    
    return m

}

func (m *manager) Submit(trainingID int) error {
    result := make(chan error, 1)
    
    m.submitCh <- submitRequest{
    	trainingID: trainingID,
    	result:     result,
    }
    
    return <-result
}

func (m *manager) run() {
    // Карта доступна только управляющей горутине.
    // Ключ — ID модели, значение — ID выполняющейся задачи обучения.
    active := make(map[int]int)
    
    for {
    	select {
    	case request := <-m.submitCh:
    		m.handleSubmit(request, active)
    
    	case result := <-m.resultCh:
    		m.handleResult(result, active)
    	}
    }
}

func (m *manager) handleSubmit(
    request submitRequest,
    active map[int]int,
) {
    t, err := m.repo.Get(request.trainingID)
    if err != nil {
        request.result <- err
        return
    }
    
    if _, exists := active[t.ModelID]; exists {
    	request.result <- ErrTrainingInProgress
    	return
    }
    
    if m.factory == nil || m.store == nil {
    	request.result <- ErrAlgorithmUnavailable
    	return
    }
    
    alg := m.factory()
    if alg == nil {
    	request.result <- ErrAlgorithmUnavailable
    	return
    }
    
    t.Status = StatusRunning
    t.Error = ""
    
    if err := m.repo.Update(t); err != nil {
    	request.result <- err
    	return
    }
    
    active[t.ModelID] = t.ID
    
    // Подтверждаем принятие задачи до запуска обучения.
    request.result <- nil
    
    go func() {
    	state, err := alg.Train([]byte(t.Text))
    
    	m.resultCh <- trainingResult{
    		trainingID: t.ID,
    		state:      state,
    		err:        err,
    	}
    }()

}

func (m *manager) handleResult(
    result trainingResult,
    active map[int]int,
) {
    t, err := m.repo.Get(result.trainingID)
    if err != nil {
        return
    }
    
    if result.err != nil {
    	t.Status = StatusFailed
    	t.Error = result.err.Error()
    } else {
    	// Сначала сохраняем обученное состояние.
    	if err := m.store.Save(algorithm.ModelKey(t.ModelID), result.state); err != nil {
    		t.Status = StatusFailed
    		t.Error = err.Error()
    	} else {
    		t.Status = StatusCompleted
    		t.Error = ""
    	}
    }
    
    if err := m.repo.Update(t); err != nil {
    	// При ошибке обновления задача может остаться в running.
    	// Для production потребуется механизм повторов.
    	return
    }
    
    if active[t.ModelID] == t.ID {
    	delete(active, t.ModelID)
    }

}
