build:
	go build scripts/generate.go
	go build scripts/download.go

clean:
	rm -f index.html

distclean: clean
	rm -rf zips mp3 generate download

update: build download
	./generate
	aws s3 cp --acl public-read --recursive mp3 s3://cbw.calledby.name/mp3/
	aws s3 cp --acl public-read index.html s3://cbw.calledby.name/

download: build
	./download