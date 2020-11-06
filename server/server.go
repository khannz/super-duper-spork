//go:generate mkdir -p ../sdslogic
//go:generate protoc --proto_path=../proto --go_out=../sdslogic --go-grpc_out=../sdslogic --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative sds.proto

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	sds "github.com/khannz/super-duper-spork/sdslogic"
	"github.com/matoous/go-nanoid"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"google.golang.org/grpc"
)

var (
	limit  = flag.IntP("limit", "l", 0, "root level limit to count quotas against")
	port   = flag.IntP("port", "p", 3009, "service port")
	prefix = flag.StringP("data-prefix", "d", "/tmp", "path to store service's files")
	re     = regexp.MustCompile(`^lbost1.(?P<Owner>[a-z]+).(?P<Quota>\d+).(?P<Id>\S+)$`)
)

type utilisatorServer struct {
	sds.UnimplementedUtilisatorServer

	// TODO: rewrite filesystem reads so that it only update map here with all data about quotes
	//existingQuotas []*sds.Quota
}

//func (s *utilisatorServer) GetUtilisation(ctx context.Context, req *sds.UtilisationRequest) (*sds.UtilisationResponse, error) {
//	//for k, quote := range s.existingQuotes {
//	//	fmt.Printf("%d and %s", k, quote)
//	//	return nil, nil
//	//}
//	log.WithFields(log.Fields{
//		"id": req.Id,
//	}).Info("GetUtilisation request")
//
//	r := &sds.UtilisationResponse{
//		Id:      req.Id,
//		Limit:   strconv.Itoa(*limit),
//		Current: countUtilisation(),
//	}
//	return r, nil
//}

func (s *utilisatorServer) AddQuota(ctx context.Context, q *sds.Quota) (*sds.QuotasSummary, error) {
	log.WithFields(log.Fields{
		"quota":      q.Size,
		"owner":      q.Owner,
		"request-id": q.Rid,
	}).Info("AddQuota request")

	n := "lbost1." + strings.ToLower(q.Owner) + "." + strconv.Itoa(int(q.Size))
	if err := writeQuotaFile(*prefix, n); err != nil {
		return nil, fmt.Errorf("cant't write QuotaFile: %s", err)
	}

	return getQuotas(*prefix, q.Rid)
}

func (s *utilisatorServer) DelQuota(ctx context.Context, q *sds.QuotaId) (*sds.QuotasSummary, error) {
	log.WithFields(log.Fields{
		"quota-id": q.Id,
	}).Info("DelQuota request")

	if delFileByID(*prefix, q.Id) != nil {
		return nil, fmt.Errorf("can't delete QuoteFile")
	}

	return getQuotas(*prefix, q.Rid)
}

func countUtilisation() string {
	return "noidea"
}

func newServer() *utilisatorServer {
	s := &utilisatorServer{}
	return s
}

func delFileByID(p, id string) error {
	f, err := os.Open(p)
	if err != nil {
		log.Errorf("failed when tried to read data-prefix dir: %s", err)
		return fmt.Errorf("can't read suffix dir")
	}
	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		log.Errorf("failed when tried to read files at data-prefix dir: %s", err)
		return fmt.Errorf("can't read files in suffix dir")
	}
	for _, file := range files {
		if len(id) == 21 && strings.HasSuffix(file.Name(), id) {
			//fmt.Println(p + "/" + file.Name())
			os.Remove(p + "/" + file.Name())
			//return nil
		}
	}
	return nil
}

func getQuotas(p, rid string) (*sds.QuotasSummary, error) {
	// TODO: here possibly re.SubexpNames() can be used
	//filenames := re.SubexpNames()

	var quotas []*sds.Quota

	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("some error here when tried to read data-prefix dir")
	}
	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, fmt.Errorf("some error here when tried to read files at data-prefix dir")
	}
	for _, file := range files {
		if re.MatchString(file.Name()) {
			s := strings.Split(file.Name(), ".")
			size, err := strconv.ParseInt(s[2], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("can't convert quota size to int")
			}
			q := &sds.Quota{
				Id:    s[3],
				Owner: s[1],
				Size:  int32(size),
			}
			quotas = append(quotas, q)
		}
	}
	return &sds.QuotasSummary{
		Quotas: quotas,
		Rid:    rid,
	}, nil
}

func writeQuotaFile(p, n string) error {
	// TODO: figure out what to do if Nanoid would be same
	// https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
	id, err := gonanoid.Nanoid()
	if err != nil {
		log.Errorf("can't make nanoid: %s", err)
		return fmt.Errorf("got some error during Nanoid generation")
	}

	r := p + "/" + n + "." + id

	if ioutil.WriteFile(r, []byte(""), 0600) != nil {
		log.WithFields(log.Fields{
			"path": r,
		}).Error("can't write QuotaFile")
		return fmt.Errorf("can't write QuotaFile %s", r)
	}
	log.WithFields(log.Fields{
		"path": r,
	}).Trace("wrote QuotaFile")
	return nil
}

//// fileExists checks if a file exists and is not a directory before we
//// try using it to prevent further errors.
//func fileExists(filename string) bool {
//	info, err := os.Stat(filename)
//	if os.IsNotExist(err) {
//		return false
//	}
//	return !info.IsDir()
//}

//func check(e error) {
//	if e != nil {
//		panic(e)
//	}
//}

func main() {
	flag.Parse()
	log.SetLevel(log.TraceLevel)

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	sds.RegisterUtilisatorServer(grpcServer, newServer())

	log.WithFields(log.Fields{
		"addr":        "localhost",
		"port":        *port,
		"limit":       *limit,
		"data-prefix": *prefix,
	}).Info("gRPC Server started")

	grpcServer.Serve(lis)
}
