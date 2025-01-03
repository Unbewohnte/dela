all: savedb clean
	mkdir -p bin && \
	 cd src && CGO_ENABLED=0 go build && mv dela ../bin && \
	 cd .. && \
	 cp -r pages bin && \
	 cp -r scripts bin && \
	 cp -r static bin

	-mv dela.db bin/

portable: clean all
	cd bin/ && cp ../COPYING . && cp ../README.md . && zip -r dela.zip * && mv dela.zip ..

savedb:
	-cp bin/dela.db .

cross: clean
	mkdir -p bin
	mkdir -p bin/dela_linux_x64
	cp -r pages bin/dela_linux_x64
	cp -r scripts bin/dela_linux_x64
	cp -r static bin/dela_linux_x64
	cp COPYING bin/dela_linux_x64
	cp README.md bin/dela_linux_x64

	mkdir -p bin/dela_windows_x64
	cp -r pages bin/dela_windows_x64
	cp -r scripts bin/dela_windows_x64
	cp -r static bin/dela_windows_x64
	cp COPYING bin/dela_windows_x64
	cp README.md bin/dela_windows_x64

	mkdir -p bin/dela_darwin_x64
	cp -r pages bin/dela_darwin_x64
	cp -r scripts bin/dela_darwin_x64
	cp -r static bin/dela_darwin_x64
	cp COPYING bin/dela_darwin_x64
	cp README.md bin/dela_darwin_x64
	
	mkdir -p bin/dela_darwin_arm64
	cp -r pages bin/dela_darwin_arm64
	cp -r scripts bin/dela_darwin_arm64
	cp -r static bin/dela_darwin_arm64
	cp COPYING bin/dela_darwin_arm64
	cp README.md bin/dela_darwin_arm64

	mkdir -p bin/dela_freebsd_x64
	cp -r pages bin/dela_freebsd_x64
	cp -r scripts bin/dela_freebsd_x64
	cp -r static bin/dela_freebsd_x64
	cp COPYING bin/dela_freebsd_x64
	cp README.md bin/dela_freebsd_x64


	cd src && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build && mv dela ../bin/dela_linux_x64
	cd src && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build && mv dela.exe ../bin/dela_windows_x64
	cd src && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build && mv dela ../bin/dela_darwin_x64
	cd src && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build && mv dela ../bin/dela_darwin_arm64
	cd src && CGO_ENABLED=0 GOOS=openbsd GOARCH=amd64 go build && mv dela ../bin/dela_freebsd_x64

clean:
	rm -rf bin
	