package internal

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"testgenerate_backend_subject/internal/app"
)

type Endpoints struct {
	GetSubjectsEndpoint   endpoint.Endpoint
	PostSubjectEndpoint   endpoint.Endpoint
	PutSubjectEndpoint    endpoint.Endpoint
	DeleteSubjectEndpoint endpoint.Endpoint
}

func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		GetSubjectsEndpoint:   MakeGetSubjectsEndpoint(s),
		PostSubjectEndpoint:   MakePostSubjectEndpoint(s),
		PutSubjectEndpoint:    MakePutSubjectEndpoint(s),
		DeleteSubjectEndpoint: MakeDeleteSubjectEndpoint(s),
	}
}

func (e Endpoints) GetSubjects(ctx context.Context) ([]app.Subject, error) {
	request := getSubjectsRequest{}
	response, err := e.GetSubjectsEndpoint(ctx, request)
	if err != nil {
		return []app.Subject{}, err
	}
	resp := response.(getSubjectsResponse)
	return resp.Subjects, resp.Err
}

func (e Endpoints) PostSubject(ctx context.Context, subject app.Subject) error {
	request := postSubjectRequest{subject}
	response, err := e.PostSubjectEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(postSubjectResponse)
	return resp.Err
}

func (e Endpoints) PutSubject(ctx context.Context, subject app.Subject) error {
	request := putSubjectRequest{subject}
	response, err := e.PutSubjectEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(putSubjectResponse)
	return resp.Err
}

func (e Endpoints) DeleteSubject(ctx context.Context, id int) error {
	request := deleteSubjectRequest{id}
	response, err := e.DeleteSubjectEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(deleteSubjectResponse)
	return resp.Err
}

// ----------------------------------------------------------------------------------------------------------------------

type getSubjectsRequest struct {
	User string
	Role string
}

type getSubjectsResponse struct {
	Subjects []app.Subject `json:"subjects,omitempty"`
	Err      error         `json:"err,omitempty"`
}

type postSubjectRequest struct {
	Subject app.Subject
}

type postSubjectResponse struct {
	Err error `json:"err,omitempty"`
}

type putSubjectRequest struct {
	Subject app.Subject
}

type putSubjectResponse struct {
	Err error `json:"err,omitempty"`
}

type deleteSubjectRequest struct {
	id int
}

type deleteSubjectResponse struct {
	Err error `json:"err,omitempty"`
}

// ----------------------------------------------------------------------------------------------------------------------
func MakeGetSubjectsEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		t, e := s.GetSubjects(ctx)
		return getSubjectsResponse{t, e}, nil
	}
}

func MakePostSubjectEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(postSubjectRequest)
		e := s.AddSubject(ctx, req.Subject)
		return postSubjectResponse{e}, nil
	}
}

func MakePutSubjectEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(putSubjectRequest)
		e := s.UpdateSubject(ctx, req.Subject)
		return putSubjectResponse{e}, nil
	}
}

func MakeDeleteSubjectEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(deleteSubjectRequest)
		e := s.DeleteSubject(ctx, req.id)
		return deleteSubjectResponse{e}, nil
	}
}
