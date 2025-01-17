package main

import (
	"fmt"
	"strings"
	"time"
)

func (d *Descriptor) genTestCacheSuccess(sb *strings.Builder) {
	fmt.Fprintf(sb, "func TestCache%sSuccess(t *testing.T) {\n", d.APIStructName())
	fmt.Fprint(sb, "\tff := &fakeFill{}\n")
	fmt.Fprintf(sb, "\tvar expect %s\n", d.ResponseTypeName())
	fmt.Fprint(sb, "\tff.Fill(&expect)\n")
	fmt.Fprintf(sb, "\tcache := &%s{\n", d.WithCacheAPIStructName())
	fmt.Fprintf(sb, "\t\tAPI: &%s{\n", d.FakeAPIStructName())
	fmt.Fprint(sb, "\t\t\tResponse: expect,\n")
	fmt.Fprint(sb, "\t\t},\n")
	fmt.Fprint(sb, "\t\tKVStore: &kvstore.Memory{},\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprintf(sb, "\tvar req %s\n", d.RequestTypeName())
	fmt.Fprint(sb, "\tff.Fill(&req)\n")
	fmt.Fprint(sb, "\tctx := context.Background()\n")
	fmt.Fprint(sb, "\tresp, err := cache.Call(ctx, req)\n")
	fmt.Fprint(sb, "\tif err != nil {\n")
	fmt.Fprint(sb, "\t\tt.Fatal(err)\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprint(sb, "\tif resp == nil {\n")
	fmt.Fprint(sb, "\t\tt.Fatal(\"expected non-nil response\")\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprint(sb, "\tif diff := cmp.Diff(expect, resp); diff != \"\" {\n")
	fmt.Fprint(sb, "\t\tt.Fatal(diff)\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprint(sb, "}\n\n")
}

func (d *Descriptor) genTestWriteCacheError(sb *strings.Builder) {
	fmt.Fprintf(sb, "func TestCache%sWriteCacheError(t *testing.T) {\n", d.APIStructName())
	fmt.Fprint(sb, "\terrMocked := errors.New(\"mocked error\")\n")
	fmt.Fprint(sb, "\tff := &fakeFill{}\n")
	fmt.Fprintf(sb, "\tvar expect %s\n", d.ResponseTypeName())
	fmt.Fprint(sb, "\tff.Fill(&expect)\n")
	fmt.Fprintf(sb, "\tcache := &%s{\n", d.WithCacheAPIStructName())
	fmt.Fprintf(sb, "\t\tAPI: &%s{\n", d.FakeAPIStructName())
	fmt.Fprint(sb, "\t\t\tResponse: expect,\n")
	fmt.Fprint(sb, "\t\t},\n")
	fmt.Fprint(sb, "\t\tKVStore: &FakeKVStore{SetError: errMocked},\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprintf(sb, "\tvar req %s\n", d.RequestTypeName())
	fmt.Fprint(sb, "\tff.Fill(&req)\n")
	fmt.Fprint(sb, "\tctx := context.Background()\n")
	fmt.Fprint(sb, "\tresp, err := cache.Call(ctx, req)\n")
	fmt.Fprint(sb, "\tif !errors.Is(err, errMocked) {\n")
	fmt.Fprint(sb, "\t\tt.Fatal(\"not the error we expected\", err)\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprint(sb, "\tif resp != nil {\n")
	fmt.Fprint(sb, "\t\tt.Fatal(\"expected nil response\")\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprint(sb, "}\n\n")
}

func (d *Descriptor) genTestFailureWithNoCache(sb *strings.Builder) {
	fmt.Fprintf(sb, "func TestCache%sFailureWithNoCache(t *testing.T) {\n", d.APIStructName())
	fmt.Fprint(sb, "\terrMocked := errors.New(\"mocked error\")\n")
	fmt.Fprint(sb, "\tff := &fakeFill{}\n")
	fmt.Fprintf(sb, "\tcache := &%s{\n", d.WithCacheAPIStructName())
	fmt.Fprintf(sb, "\t\tAPI: &%s{\n", d.FakeAPIStructName())
	fmt.Fprint(sb, "\t\t\tErr: errMocked,\n")
	fmt.Fprint(sb, "\t\t},\n")
	fmt.Fprint(sb, "\t\tKVStore: &kvstore.Memory{},\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprintf(sb, "\tvar req %s\n", d.RequestTypeName())
	fmt.Fprint(sb, "\tff.Fill(&req)\n")
	fmt.Fprint(sb, "\tctx := context.Background()\n")
	fmt.Fprint(sb, "\tresp, err := cache.Call(ctx, req)\n")
	fmt.Fprint(sb, "\tif !errors.Is(err, errMocked) {\n")
	fmt.Fprint(sb, "\t\tt.Fatal(\"not the error we expected\", err)\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprint(sb, "\tif resp != nil {\n")
	fmt.Fprint(sb, "\t\tt.Fatal(\"expected nil response\")\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprint(sb, "}\n\n")
}

func (d *Descriptor) genTestFailureWithPreviousCache(sb *strings.Builder) {
	// This works for both caching policies.
	fmt.Fprintf(sb, "func TestCache%sFailureWithPreviousCache(t *testing.T) {\n", d.APIStructName())
	fmt.Fprint(sb, "\tff := &fakeFill{}\n")
	fmt.Fprintf(sb, "\tvar expect %s\n", d.ResponseTypeName())
	fmt.Fprint(sb, "\tff.Fill(&expect)\n")
	fmt.Fprintf(sb, "\tfakeapi := &%s{\n", d.FakeAPIStructName())
	fmt.Fprint(sb, "\t\tResponse: expect,\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprintf(sb, "\tcache := &%s{\n", d.WithCacheAPIStructName())
	fmt.Fprint(sb, "\t\tAPI: fakeapi,\n")
	fmt.Fprint(sb, "\t\tKVStore: &kvstore.Memory{},\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprintf(sb, "\tvar req %s\n", d.RequestTypeName())
	fmt.Fprint(sb, "\tff.Fill(&req)\n")
	fmt.Fprint(sb, "\tctx := context.Background()\n")
	fmt.Fprint(sb, "\t// first pass with no error at all\n")
	fmt.Fprint(sb, "\t// use a separate scope to be sure we avoid mistakes\n")
	fmt.Fprint(sb, "\t{\n")
	fmt.Fprint(sb, "\t\tresp, err := cache.Call(ctx, req)\n")
	fmt.Fprint(sb, "\t\tif err != nil {\n")
	fmt.Fprint(sb, "\t\t\tt.Fatal(err)\n")
	fmt.Fprint(sb, "\t\t}\n")
	fmt.Fprint(sb, "\t\tif resp == nil {\n")
	fmt.Fprint(sb, "\t\t\tt.Fatal(\"expected non-nil response\")\n")
	fmt.Fprint(sb, "\t\t}\n")
	fmt.Fprint(sb, "\t\tif diff := cmp.Diff(expect, resp); diff != \"\" {\n")
	fmt.Fprint(sb, "\t\t\tt.Fatal(diff)\n")
	fmt.Fprint(sb, "\t\t}\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprint(sb, "\t// second pass with failure\n")
	fmt.Fprint(sb, "\terrMocked := errors.New(\"mocked error\")\n")
	fmt.Fprint(sb, "\tfakeapi.Err = errMocked\n")
	fmt.Fprint(sb, "\tfakeapi.Response = nil\n")
	fmt.Fprint(sb, "\tresp2, err := cache.Call(ctx, req)\n")
	fmt.Fprint(sb, "\tif err != nil {\n")
	fmt.Fprint(sb, "\t\tt.Fatal(err)\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprint(sb, "\tif resp2 == nil {\n")
	fmt.Fprint(sb, "\t\tt.Fatal(\"expected non-nil response\")\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprint(sb, "\tif diff := cmp.Diff(expect, resp2); diff != \"\" {\n")
	fmt.Fprint(sb, "\t\tt.Fatal(diff)\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprint(sb, "}\n\n")
}

func (d *Descriptor) genTestSetcacheWithEncodeError(sb *strings.Builder) {
	fmt.Fprintf(sb, "func TestCache%sSetcacheWithEncodeError(t *testing.T) {\n", d.APIStructName())
	fmt.Fprint(sb, "\tff := &fakeFill{}\n")
	fmt.Fprint(sb, "\terrMocked := errors.New(\"mocked error\")\n")
	fmt.Fprintf(sb, "\tvar in []%s\n", d.CacheEntryName())
	fmt.Fprint(sb, "\tff.Fill(&in)\n")
	fmt.Fprintf(sb, "\tcache := &%s{\n", d.WithCacheAPIStructName())
	fmt.Fprint(sb, "\t\tGobCodec: &FakeCodec{EncodeErr: errMocked},\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprintf(sb, "\terr := cache.setcache(in)\n")
	fmt.Fprint(sb, "\tif !errors.Is(err, errMocked) {\n")
	fmt.Fprint(sb, "\t\tt.Fatal(\"not the error we expected\", err)\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprint(sb, "}\n\n")
}

func (d *Descriptor) genTestReadCacheNotFound(sb *strings.Builder) {
	if fields := d.StructFields(d.Request); len(fields) <= 0 {
		// this test cannot work when there are no fields in the
		// request because we will always find a match.
		// TODO(bassosimone): how to avoid having uncovered code?
		return
	}
	fmt.Fprintf(sb, "func TestCache%sReadCacheNotFound(t *testing.T) {\n", d.APIStructName())
	fmt.Fprint(sb, "\tff := &fakeFill{}\n")
	fmt.Fprintf(sb, "\tvar incache []%s\n", d.CacheEntryName())
	fmt.Fprint(sb, "\tff.Fill(&incache)\n")
	fmt.Fprintf(sb, "\tcache := &%s{\n", d.WithCacheAPIStructName())
	fmt.Fprint(sb, "\t\tKVStore: &kvstore.Memory{},\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprintf(sb, "\terr := cache.setcache(incache)\n")
	fmt.Fprintf(sb, "\tif err != nil {\n")
	fmt.Fprintf(sb, "\t\tt.Fatal(err)\n")
	fmt.Fprintf(sb, "\t}\n")
	fmt.Fprintf(sb, "\tvar req %s\n", d.RequestTypeName())
	fmt.Fprint(sb, "\tff.Fill(&req)\n")
	fmt.Fprintf(sb, "\tout, err := cache.readcache(req)\n")
	fmt.Fprint(sb, "\tif !errors.Is(err, errCacheNotFound) {\n")
	fmt.Fprint(sb, "\t\tt.Fatal(\"not the error we expected\", err)\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprint(sb, "\tif out != nil {\n")
	fmt.Fprint(sb, "\t\tt.Fatal(\"expected nil here\")\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprint(sb, "}\n\n")
}

func (d *Descriptor) genTestWriteCacheDuplicate(sb *strings.Builder) {
	fmt.Fprintf(sb, "func TestCache%sWriteCacheDuplicate(t *testing.T) {\n", d.APIStructName())
	fmt.Fprint(sb, "\tff := &fakeFill{}\n")
	fmt.Fprintf(sb, "\tvar req %s\n", d.RequestTypeName())
	fmt.Fprint(sb, "\tff.Fill(&req)\n")
	fmt.Fprintf(sb, "\tvar resp1 %s\n", d.ResponseTypeName())
	fmt.Fprint(sb, "\tff.Fill(&resp1)\n")
	fmt.Fprintf(sb, "\tvar resp2 %s\n", d.ResponseTypeName())
	fmt.Fprint(sb, "\tff.Fill(&resp2)\n")
	fmt.Fprintf(sb, "\tcache := &%s{\n", d.WithCacheAPIStructName())
	fmt.Fprint(sb, "\t\tKVStore: &kvstore.Memory{},\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprintf(sb, "\terr := cache.writecache(req, resp1)\n")
	fmt.Fprintf(sb, "\tif err != nil {\n")
	fmt.Fprintf(sb, "\t\tt.Fatal(err)\n")
	fmt.Fprintf(sb, "\t}\n")
	fmt.Fprintf(sb, "\terr = cache.writecache(req, resp2)\n")
	fmt.Fprintf(sb, "\tif err != nil {\n")
	fmt.Fprintf(sb, "\t\tt.Fatal(err)\n")
	fmt.Fprintf(sb, "\t}\n")
	fmt.Fprintf(sb, "\tout, err := cache.readcache(req)\n")
	fmt.Fprint(sb, "\tif err != nil {\n")
	fmt.Fprint(sb, "\t\tt.Fatal(err)\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprint(sb, "\tif out == nil {\n")
	fmt.Fprint(sb, "\t\tt.Fatal(\"expected non-nil here\")\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprint(sb, "\tif diff := cmp.Diff(resp2, out); diff != \"\" {\n")
	fmt.Fprint(sb, "\t\tt.Fatal(diff)\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprint(sb, "}\n\n")
}

func (d *Descriptor) genTestCachSizeLimited(sb *strings.Builder) {
	if fields := d.StructFields(d.Request); len(fields) <= 0 {
		// this test cannot work when there are no fields in the
		// request because we will always find a match.
		// TODO(bassosimone): how to avoid having uncovered code?
		return
	}
	fmt.Fprintf(sb, "func TestCache%sCacheSizeLimited(t *testing.T) {\n", d.APIStructName())
	fmt.Fprint(sb, "\tff := &fakeFill{}\n")
	fmt.Fprintf(sb, "\tcache := &%s{\n", d.WithCacheAPIStructName())
	fmt.Fprint(sb, "\t\tKVStore: &kvstore.Memory{},\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprintf(sb, "\tvar prev int\n")
	fmt.Fprintf(sb, "\tfor {\n")
	fmt.Fprintf(sb, "\t\tvar req %s\n", d.RequestTypeName())
	fmt.Fprint(sb, "\t\tff.Fill(&req)\n")
	fmt.Fprintf(sb, "\t\tvar resp %s\n", d.ResponseTypeName())
	fmt.Fprint(sb, "\t\tff.Fill(&resp)\n")
	fmt.Fprintf(sb, "\t\terr := cache.writecache(req, resp)\n")
	fmt.Fprintf(sb, "\t\tif err != nil {\n")
	fmt.Fprintf(sb, "\t\t\tt.Fatal(err)\n")
	fmt.Fprintf(sb, "\t\t}\n")
	fmt.Fprintf(sb, "\t\tout, err := cache.getcache()\n")
	fmt.Fprint(sb, "\t\tif err != nil {\n")
	fmt.Fprint(sb, "\t\t\tt.Fatal(err)\n")
	fmt.Fprint(sb, "\t\t}\n")
	fmt.Fprint(sb, "\t\tif len(out) > prev {\n")
	fmt.Fprint(sb, "\t\t\tprev = len(out)\n")
	fmt.Fprint(sb, "\t\t\tcontinue\n")
	fmt.Fprint(sb, "\t\t}\n")
	fmt.Fprint(sb, "\t\tbreak\n")
	fmt.Fprint(sb, "\t}\n")
	fmt.Fprint(sb, "}\n\n")
}

// GenCachingTestGo generates caching_test.go.
func GenCachingTestGo(file string) {
	var sb strings.Builder
	fmt.Fprint(&sb, "// Code generated by go generate; DO NOT EDIT.\n")
	fmt.Fprintf(&sb, "// %s\n\n", time.Now())
	fmt.Fprint(&sb, "package ooapi\n\n")
	fmt.Fprintf(&sb, "//go:generate go run ./internal/generator -file %s\n\n", file)
	fmt.Fprint(&sb, "import (\n")
	fmt.Fprint(&sb, "\t\"context\"\n")
	fmt.Fprint(&sb, "\t\"errors\"\n")
	fmt.Fprint(&sb, "\t\"testing\"\n")
	fmt.Fprint(&sb, "\n")
	fmt.Fprint(&sb, "\t\"github.com/google/go-cmp/cmp\"\n")
	fmt.Fprint(&sb, "\t\"github.com/ooni/probe-cli/v3/internal/kvstore\"\n")
	fmt.Fprint(&sb, "\t\"github.com/ooni/probe-cli/v3/internal/ooapi/apimodel\"\n")
	fmt.Fprint(&sb, ")\n")
	for _, desc := range Descriptors {
		if desc.CachePolicy == CacheNone {
			continue
		}
		desc.genTestCacheSuccess(&sb)
		desc.genTestWriteCacheError(&sb)
		desc.genTestFailureWithNoCache(&sb)
		desc.genTestFailureWithPreviousCache(&sb)
		desc.genTestSetcacheWithEncodeError(&sb)
		desc.genTestReadCacheNotFound(&sb)
		desc.genTestWriteCacheDuplicate(&sb)
		desc.genTestCachSizeLimited(&sb)
	}
	writefile(file, &sb)
}
