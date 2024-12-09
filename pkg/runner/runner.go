package runner

// import (
//     "sync"
// )

// func ExecuteTasks(cfg *Config, logger *Logger) {
//     var wg sync.WaitGroup

//     for _, task := range cfg.Images {
//         wg.Add(1)
//         go func(task ImageConfig) {
//             defer wg.Done()
//             logger.Info("Executing task for image: " + task.Dockerfile)
//             // Do work here
//         }(task)
//     }

//     wg.Wait()
//     logger.Info("All tasks completed")
// }
