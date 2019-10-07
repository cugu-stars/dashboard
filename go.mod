module github.com/cugu/dashboard

go 1.13

require (
	github.com/dustin/go-humanize v1.0.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/narqo/go-badge v0.0.0-20190124110329-d9415e4e1e9f
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/xanzy/go-gitlab v0.20.1
	golang.org/x/image v0.0.0-20190910094157-69e4b8554b2a // indirect
	gopkg.in/russross/blackfriday.v2 v2.0.0-00010101000000-000000000000
	gopkg.in/yaml.v2 v2.2.4
)

replace gopkg.in/russross/blackfriday.v2 => github.com/russross/blackfriday/v2 v2.0.1
