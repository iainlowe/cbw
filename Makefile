build:
	go build scripts/generate.go
	go build scripts/download.go

clean:
	rm -f index.html

distclean: clean
	rm -rf zips mp3 generate download

update: download
	./generate
	s3sync -r -md5 -acl public-read mp3 cbw.calledby.name/mp3
	s3sync -md5 -acl public-read index.html cbw.calledby.name

download:
	./download