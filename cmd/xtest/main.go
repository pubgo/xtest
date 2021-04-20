package main

import (
	"fmt"
	"os"

	"github.com/dave/jennifer/jen"
	"github.com/pubgo/xtest/internal"
)

func main() {
	gt, err := internal.GenerateTests(os.Args[1], nil)
	if err != nil {
		panic(err)
	}

	for _, g := range gt[:] {
		fmt.Println(g.Path)
		for _, fn := range g.Functions {
			if fn.Receiver == nil {
				fmt.Printf("%#v\n", jen.Func().Id("Test"+fn.Name).Params(jen.Id("t").Id("*").Qual("testing", "T")).BlockFunc(func(g *jen.Group) {
					g.Id("fn").Op(":=").Qual("github.com/pubgo/xtest", "TestFuncWith").CallFunc(func(g *jen.Group) {
						g.Func().ParamsFunc(func(g *jen.Group) {
							for _, p := range fn.Parameters {
								g.Id(p.Name).Id(p.Type.String())
							}
						}).BlockFunc(func(g *jen.Group) {
							if len(fn.Results) == 0 {
								g.Id(fn.Name).CallFunc(func(g *jen.Group) {
									for _, p := range fn.Parameters {
										if p.Type.IsVariadic {
											g.Id(p.Name).Id("...")
										} else {
											g.Id(p.Name)
										}
									}
								})
							} else {
								g.ListFunc(func(g *jen.Group) {
									for i := range fn.Results {
										g.Id(fmt.Sprintf("r%d", i+1))
									}
								}).Op(":=").Id(fn.Name).CallFunc(func(g *jen.Group) {
									for _, p := range fn.Parameters {
										if p.Type.IsVariadic {
											g.Id(p.Name).Id("...")
										} else {
											g.Id(p.Name)
										}
									}
								})
							}
						})
					})
					g.Qual("fn", "In").Call()
					g.Qual("fn", "Do").Call()
				}))
			} else {
				typ := fn.Receiver.Type.Value + "Fixture"
				fmt.Printf("%#v\n", jen.Type().Id(typ).Struct(
					jen.Id("*").Qual("github.com/smartystreets/gunit", "Fixture"),
					jen.Id("unit").Id("*").Id(fn.Receiver.Type.Value),
				))

				fmt.Printf("%#v\n", jen.Func().Params(jen.Id(fn.Receiver.Name).Id("*").Id(typ)).Id("Test"+fn.Name).ParamsFunc(func(g *jen.Group) {
					for _, p := range fn.Parameters {
						g.Id(p.Name).Id(p.Type.String())
					}
				}).BlockFunc(func(g *jen.Group) {
					g.Id("fn").Op(":=").Qual("github.com/pubgo/xtest", "TestFuncWith").CallFunc(func(g *jen.Group) {
						g.Func().ParamsFunc(func(g *jen.Group) {
							for _, p := range fn.Parameters {
								g.Id(p.Name).Id(p.Type.String())
							}
						}).BlockFunc(func(g *jen.Group) {
							if len(fn.Results) == 0 {
								g.Id(fn.Name).CallFunc(func(g *jen.Group) {
									for _, p := range fn.Parameters {
										if p.Type.IsVariadic {
											g.Id(p.Name).Id("...")
										} else {
											g.Id(p.Name)
										}
									}
								})
							} else {
								g.ListFunc(func(g *jen.Group) {
									for i := range fn.Results {
										g.Id(fmt.Sprintf("r%d", i+1))
									}
								}).Op(":=").Id(fn.Name).CallFunc(func(g *jen.Group) {
									for _, p := range fn.Parameters {
										if p.Type.IsVariadic {
											g.Id(p.Name).Id("...")
										} else {
											g.Id(p.Name)
										}
									}
								})
							}
						})
					})
					g.Qual("fn", "In").Call()
					g.Qual("fn", "Do").Call()
				}))
			}

			//_dt, _ := json.MarshalIndent(fn, " ", "\t")
			//fmt.Println(string(_dt))
		}
	}
}
