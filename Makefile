all:
	mkdir -p bin && \
	 cd src && CGO_ENABLED=0 go build && mv dela ../bin && \
	 cd .. && \
	 cp -r pages bin && \
	 cp -r scripts bin && \
	 cp -r static bin


clean:
	rm -rf bin
	