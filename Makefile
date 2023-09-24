all:
	mkdir -p bin && \
	 cd src && CGO_ENABLED=0 go build && mv dela ../bin && \
	 cd .. && \
	 cp -r pages bin && \
	 cp -r scripts bin && \
	 cp -r static bin

portable: clean all
	cd bin/ && cp ../COPYING . && cp ../README.md . && zip -r dela.zip * && mv dela.zip ..

clean:
	rm -rf bin
	