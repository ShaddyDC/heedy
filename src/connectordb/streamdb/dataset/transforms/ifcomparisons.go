package transforms

import (
	"connectordb/streamdb/datastream"
	"errors"
)

//BooleanFilterTransform is a transform that returns the datapoint if the internal transform is true,
//and returns nil if the internal transform is false
type BooleanFilterTransform struct {
	booltransform DatapointTransform
}

func (t BooleanFilterTransform) Transform(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
	if dp == nil {
		return nil, nil
	}
	tdp, err = t.booltransform.Transform(dp)
	if err != nil {
		return nil, err
	}
	if tdp == nil {
		return nil, errors.New("Filter received null value")
	}
	filter, ok := tdp.Data.(bool)
	if !ok {
		return nil, errors.New("Filter did not return boolean value")
	}
	if filter {
		return dp, nil
	}
	//The datapoint was filtered
	return nil, nil
}

func IfEq(args []string) (DatapointTransform, error) {
	t, err := Eq(args)
	return BooleanFilterTransform{t}, err
}

func IfLt(args []string) (DatapointTransform, error) {
	t, err := Lt(args)
	return BooleanFilterTransform{t}, err
}

func IfGt(args []string) (DatapointTransform, error) {
	t, err := Gt(args)
	return BooleanFilterTransform{t}, err
}

func IfGte(args []string) (DatapointTransform, error) {
	t, err := Gte(args)
	return BooleanFilterTransform{t}, err
}

func IfLte(args []string) (DatapointTransform, error) {
	t, err := Lte(args)
	return BooleanFilterTransform{t}, err
}

//OrTransform takes multiple transforms and returns true if any one of them returns true
type OrTransform struct {
	booltransforms []DatapointTransform
}

func (t OrTransform) Transform(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
	if dp == nil {
		return nil, nil
	}
	result := CopyDatapoint(dp)

	for i := 0; i < len(t.booltransforms); i++ {
		tdp, err = t.booltransforms[i].Transform(dp)
		if err != nil {
			return nil, err
		}
		if tdp == nil {
			return nil, errors.New("or: received null value")
		}
		filter, ok := tdp.Data.(bool)
		if !ok {
			return nil, errors.New("or: statement did not return boolean value")
		}
		if filter {
			result.Data = true
			return result, nil
		}
	}
	result.Data = false
	return result, nil
}

//Given pipelines, or is exactly like an or statement
func Or(args []string) (DatapointTransform, error) {
	if len(args) == 0 {
		return nil, errors.New("or: not enough arguments")
	}
	pipelines := make([]DatapointTransform, 0, len(args))

	for i := 0; i < len(args); i++ {
		tpipe, err := NewTransformPipeline(args[i])
		if err != nil {
			return nil, err
		}
		pipelines = append(pipelines, tpipe)
	}
	return &OrTransform{pipelines}, nil
}

//If filters on its arguments - where each argument is 'or'
func If(args []string) (DatapointTransform, error) {
	t, err := Or(args)
	return BooleanFilterTransform{t}, err
}
