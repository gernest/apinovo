DEFAULT_POSTGRES_CONN	:=postgres://postgres:postgres@localhost/apinova?sslmode=disable
DEFAULT_DIALECT			:=postgres

ifeq "$(origin DB_CONN)" "undefined"
DB_CONN=$(DEFAULT_POSTGRES_CONN)
endif

test#:
	@DB_CONN=$(DB_CONN) go test -v -cover
