# Module Usage Guide: Database Stored Procedure Execution
## Prerequisites
- You should have a working knowledge of the Go programming language.
- You need to have the SQLx library installed in your Go environment. You can install it using

    ```sh
    go get github.com/jmoiron/sqlx
    ```
## Installation
```sh
go get github.com/ignaciocaff/oraclesp
```
## Function Overview

```go
Execute(procedureName string, result interface{}, args ...interface{}) er
```
- **procedureName**: The name of the stored procedure to execute.
- **result**: A pointer to the structure or object where the results will be mapped.
- **args**: Variadic arguments representing the parameters required by the stored procedure.

## Usage Example

```go
import (
	"github.com/ignaciocaff/oraclesp"
)
type Employee struct {
    Id        int       `oracle:"ID"`
    FirstName string    `oracle:"FIRST_NAME"`
    LastName  string    `oracle:"LAST_NAME"`
    Birthdate time.Time `oracle:"BIRTHDATE"`
}

func FunctionName(param1, param2 int) (Employee, error) {
    var entity Employee
    err := oraclesp.Execute("PACKAGE_NAME.STORE_PROCEDURE_NAME", &res, param1, param2)
    if err != nil {
        return nil, err
    }
    return res, nil
}
```
#### The name following the "oracle" tag must match the alias of the column belonging to the output cursor in Oracle.

```go
oracle:"FIRST_NAME"
```


