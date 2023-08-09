# Module Usage Guide: Database Stored Procedure Execution

## Objective
The objective of this library is to facilitate the execution of stored procedures in a database using the Go programming language. It specifically targets Oracle databases and aims to simplify the process of invoking stored procedures and mapping their results to Go structures or objects.
Additionally, the library utilizes the driver **github.com/sijms/go-ora** for Oracle database connectivity. One noteworthy feature is that it eliminates the need to install any version of the Oracle Instant Client. This characteristic streamlines the setup process and reduces the external dependencies required to connect to Oracle databases.


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


