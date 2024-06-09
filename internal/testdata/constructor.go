package testdata

import "sync"

// START unexported struct
type unexported struct {
	a    string
	b    int
	c    bool
	d, e *sync.Mutex
} // END

/* EXPECT
// START unexported struct
type unexported struct {
	a    string
	b    int
	c    bool
	d, e *sync.Mutex
} // END


func newUnexported(
	a string,
	b int,
	c bool,
	d *sync.Mutex,
	e *sync.Mutex,
) unexported {
	return unexported{
		a: a,
		b: b,
		c: c,
		d: d,
		e: e,
	}
}
*/

// START exported struct
type Exported struct {
	A string
} // END

/* EXPECT
// START exported struct
type Exported struct {
	A string
} // END


func NewExported(
	a string,
) Exported {
	return Exported{
		A: a,
	}
}
*/

// START embedded struct
type Embedded struct {
	Exported
	B int
} // END

/* EXPECT
// START embedded struct
type Embedded struct {
	Exported
	B int
} // END


func NewEmbedded(
	exported Exported,
	b int,
) Embedded {
	return Embedded{
		Exported: exported,
		B:        b,
	}
}
*/
