package algorithm

type Trainer interface {
	Train(data []byte) ([]byte, error)
}

type Predictor interface {
	Predict(state []byte, input []byte) ([]byte, error)
}

type Algorithm interface {
	Trainer
	Predictor
}

type Factory func() Algorithm
