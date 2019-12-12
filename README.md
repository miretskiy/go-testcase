### Generate table driven tests easily.

Table driven tests are ubiquitous in Go.
Unfortunately, if you would like to have a test case
that represents a cartesian product of all of the settings,
you're left with large amount of boiler plat:

```gotemplate
func Test(t *testing.T) {
  testCases := []struct{
    a bool
    b bool
    c bool
  }{
    {false, false, false},
    {false, false, true},
    ...
  }
  for _, tc := range testCases {
     // Execute test case tc
  }
}
``` 

Often enough, the settings are complex (perhaps `map[string][]int`), and
writing cartesian product by hand results in fairly large amount of boiler plate.
While writing verbose code is not a bad thing in Go, I believe that writing
the boiler plate for the sake of boilerplate is a *bad* thing.

Usually, at this point, you would abandon table driven tests, and switch to
using nested for loops.

Enter this library.  It allows you to generate test cases easily:
  * Define a struct representing your test case
  * Annotate fields you care about with *tc* tag.  The value of the
    tag is a JSON array of values you wish to assign to the field.
    Type checking is performed to ensure JSON values can be converted
    to the field type.
  * Use test generator to generate your test cases for you.           
                                           
```gotemplate
func Test(t *testing.T) {
  tc := &struct {
    enable bool  `tc:"[true, false]"`
    i      int   `tc:"[3, 2, 1]"`
    arr    []int `tc:"[[1,2,3], [9,8,7]]"`
  }
  for gen := GenerateTestCases(t, &tc); gen.Next(); {
    // Execute test case tc: all values will be initialized
  }
}
``` 

See unit test for more examples.  Have fun.
