module github.com/ethodomingues/slow

go 1.19

require golang.org/x/exp v0.0.0-20221111204811-129d8d6c17ab

require (
	github.com/ethodomingues/c3po v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/ethodomingues/c3po => ../c3po
