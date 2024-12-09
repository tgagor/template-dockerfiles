package parser

// import (
//     // "log"
//     "text/template"
// )

// func Run(cfg *Config, logger *Logger) error {
//     tmpl, err := template.ParseFiles(cfg.Dockerfile)
//     if err != nil {
//         return err
//     }

//     for name, img := range cfg.Images {
//         logger.Info("Building image: " + name)
//         // Render templates using variables
//         err := tmpl.ExecuteTemplate(os.Stdout, img.Dockerfile, img.Variables)
//         if err != nil {
//             return err
//         }
//     }
//     return nil
// }
