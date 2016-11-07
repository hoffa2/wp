# Golang worker pool
This is an implementation of a "generic" worker pool.
You can define pools to call a  function with your own
defined input argument.


## Simple Example
```go

func jobfunc(args interface{}) {
    s := args.(string)
    fmt.Println(s)
}


pool := NewPool(100, jobfunc)
pool.Start()
pool.Add("Hello workerpool")

```

## Interface example
```go

type Example struct {
    results chan float
}

type argument struct {
    a float
    b float
}

func (e *Example) Jobfunc(args interface{}) {
    ex := args.(argument)
    e.results <- ex.a / ex.b
}

e := Example{results: make(chan float, 100)}

pool := NewPool(100, e.Jobfunc)
pool.Start()
pool.Add(argument{a:10, b:20})

fmt.Println(<-e.results)
```

