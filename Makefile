
protoc:
	cd proto && protoc --go_out=../pb --go_opt=paths=source_relative --go-grpc_out=../pb --go-grpc_opt=paths=source_relative *.proto && cd ..

gen:
	cd cert; bash gen.sh; cd ..