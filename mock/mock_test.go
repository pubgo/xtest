package mock

//
//func (t *xtestFixture) TestMockRegister() {
//	fn := TestFuncWith(func(fns ...interface{}) {
//		defer xerror.Resp(func(err xerror.XErr) {
//			switch err := err.Unwrap(); err {
//			case ErrParamIsNil, ErrParamTypeNotFunc:
//			default:
//				panic(err)
//			}
//		})
//
//		MockRegister(fns...)
//	})
//	fn.In(
//		nil,
//		"hello",
//		func() {},
//	)
//	fn.Do()
//}
