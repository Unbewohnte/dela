all: clean
	mkdir -p bin && \
	 cd src && CGO_ENABLED=0 go build && mv dela ../bin && \
	 cd .. && \
	 cp -r pages bin && \
	 cp -r scripts bin && \
	 cp -r static bin

portable: clean all
	cd bin/ && cp ../COPYING . && cp ../README.md . && zip -r dela.zip * && mv dela.zip ..

cross: clean
	mkdir -p bin
	mkdir -p bin/dela_linux_x64
	mkdir -p bin/dela_linux_x32
	mkdir -p bin/dela_windows_x64
	mkdir -p bin/dela_windows_arm64
	mkdir -p bin/dela_darwin_x64
	mkdir -p bin/dela_darwin_arm64
	mkdir -p bin/dela_freebsd_x64
	mkdir -p bin/dela_freebsd_arm64



	cd src && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build && mv dela ../bin/dela_linux_x64
	cd src && CGO_ENABLED=0 GOOS=linux GOARCH=386 go build && mv dela ../bin/dela_linux_x32
	cd src && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build && mv dela.exe ../bin/dela_windows_x64
	cd src && CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build && mv dela.exe ../bin/dela_windows_arm64
	cd src && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build && mv dela ../bin/dela_darwin_x64
	cd src && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build && mv dela ../bin/dela_darwin_arm64
	cd src && CGO_ENABLED=0 GOOS=openbsd GOARCH=amd64 go build && mv dela ../bin/dela_freebsd_x64
	cd src && CGO_ENABLED=0 GOOS=openbsd GOARCH=arm64 go build && mv dela ../bin/dela_freebsd_arm64
	

	mkdir -p bin/includes

	cp -r pages bin/includes
	cp -r scripts bin/includes
	cp -r static bin/includes
	cp COPYING bin/includes
	cp LICENSE* bin/includes
	cp README.md bin/includes


clean:
	rm -rf bin
	