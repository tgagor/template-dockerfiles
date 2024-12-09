## Known bugs

- Bug: Labeling of Docker images happen during build time, which in parallel mode would result in inconsistent tagging (depending on which task would finish the last). I need to split build of images from tagging them.
  Workaround: If you rely on labels in order, don't use parallel mode.

## To do

- Feature: Template not only Dockerfiles, but also other files with `*.jinja` or `*.jinja2` extensions. This could allow to generate different scripts based on image type, but would require to isolate the complete build context in separate directory per building thread to avoid collisions. On big projects this could significantly rise the build time.
- Feature: Add Python Pip package for easy installation and usage.
- Feature: Add test for multiple Python version to ensure that app can work on them.
- Idea: Jinja is currently rather exotic type of templating, much more popular is Go Lang template. It's also much simpler to make Go project compatible with different platforms as it have not many dependencies. I might consider rewriting project to Go Lang.
