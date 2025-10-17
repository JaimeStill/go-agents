package processing

type ProgressFunc[T any] func(completed, total int, result T)
