package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
	"github.com/dagger/dagger/cmd/codegen/generator"
	gogenerator "github.com/dagger/dagger/cmd/codegen/generator/go"
	typescriptgenerator "github.com/dagger/dagger/cmd/codegen/generator/typescript"
	"github.com/dagger/dagger/cmd/codegen/introspection"
)

func Generate(ctx context.Context, cfg generator.Config, dag *dagger.Client) (err error) {
	logsW := os.Stdout

	// if cfg.ModuleName == "foo" {
	// 	raw, err := exec.Command("sh", "-c", `apk add wget && wget https://github.com/jqlang/jq/releases/download/jq-1.7.1/jq-linux-arm64 && mv jq-linux-arm64 /usr/local/bin/jq && chmod +x /usr/local/bin/jq`).CombinedOutput()
	// 	if err != nil {
	// 		return fmt.Errorf("REAL ERROR 1 FROM GENERATE %s", string(raw))
	// 	}

	// 	raw, err = exec.Command("sh", "-c", "cat /schema.json | jq | grep -i skip").CombinedOutput()
	// 	if err != nil {
	// 		return fmt.Errorf("REAL ERROR 2 FROM GENERATE %s", string(raw))
	// 	}

	// 	if strings.Contains(strings.ToLower(string(raw)), "skiptparse") {
	// 		return fmt.Errorf("ERROR FROM GENERATE -> \n%s", string(raw))
	// 	}
	// }

	if cfg.ModuleName != "" {
		fmt.Fprintf(logsW, "generating %s module: %s\n", cfg.Lang, cfg.ModuleName)
	} else {
		fmt.Fprintf(logsW, "generating %s SDK client\n", cfg.Lang)
	}

	var introspectionSchema *introspection.Schema
	var introspectionSchemaVersion string
	if cfg.IntrospectionJSON != "" {
		var resp introspection.Response
		if err := json.Unmarshal([]byte(cfg.IntrospectionJSON), &resp); err != nil {
			return fmt.Errorf("unmarshal introspection json: %w", err)
		}
		introspectionSchema = resp.Schema
		introspectionSchemaVersion = resp.SchemaVersion
	} else {
		introspectionSchema, introspectionSchemaVersion, err = generator.Introspect(ctx, dag)
		if err != nil {
			return err
		}
	}

	for ctx.Err() == nil {
		generated, err := generate(ctx, introspectionSchema, introspectionSchemaVersion, cfg)
		if err != nil {
			return err
		}

		if err := generator.Overlay(ctx, logsW, generated.Overlay, cfg.OutputDir); err != nil {
			return fmt.Errorf("failed to overlay generated code: %w", err)
		}

		for _, cmd := range generated.PostCommands {
			cmd.Dir = cfg.OutputDir
			if cfg.ModuleName != "" {
				cmd.Dir = filepath.Join(cfg.OutputDir, cfg.ModuleContextPath)
			}
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			fmt.Fprintln(logsW, "running post-command:", strings.Join(cmd.Args, " "))
			err := cmd.Run()
			if err != nil {
				fmt.Fprintln(logsW, "post-command failed:", err)
				return err
			}
		}

		if !generated.NeedRegenerate {
			fmt.Fprintln(logsW, "done!")
			break
		}

		fmt.Fprintln(logsW, "needs another pass...")
	}

	return ctx.Err()
}

func generate(ctx context.Context, introspectionSchema *introspection.Schema, introspectionSchemaVersion string, cfg generator.Config) (*generator.GeneratedState, error) {
	generator.SetSchemaParents(introspectionSchema)

	var gen generator.Generator
	switch cfg.Lang {
	case generator.SDKLangGo:
		gen = &gogenerator.GoGenerator{
			Config: cfg,
		}
	case generator.SDKLangTypeScript:
		gen = &typescriptgenerator.TypeScriptGenerator{
			Config: cfg,
		}

	default:
		sdks := []string{
			string(generator.SDKLangGo),
			string(generator.SDKLangTypeScript),
		}
		return nil, fmt.Errorf("use target SDK language: %s: %w", sdks, generator.ErrUnknownSDKLang)
	}

	return gen.Generate(ctx, introspectionSchema, introspectionSchemaVersion)
}
