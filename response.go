package afast

type Response struct {
	Status int
}

func (r *Response) ToBytes() []byte {
	if r.Status == 0 {
		r.Status = 200
	}
	return []byte("response")
}
