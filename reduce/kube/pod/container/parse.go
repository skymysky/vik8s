package container

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/refs"
	v1 "k8s.io/api/core/v1"
)

func container(d *config.Directive, pod *v1.PodSpec) v1.Container {
	asserts.ArgsRange(d, 2, 3)

	c := v1.Container{
		Name: d.Args[0], Image: d.Args[1],
	}

	if ipp := utils.Index(d.Args, 2); ipp != "" {
		c.ImagePullPolicy = v1.PullPolicy(ipp)
	}

	for _, body := range d.Body {
		switch body.Name {
		default:
			refs.UnmarshalItem(&c, body)

		case "command":
			c.Command = body.Args
		case "args":
			c.Args = body.Args

		case "port":
			asserts.ArgsMin(body, 1)
			if len(body.Args) == 1 {
				c.Ports = append(c.Ports, portParse("", body.Args[0]))
			} else {
				c.Ports = append(c.Ports, portParse(body.Args[0], body.Args[1]))
			}
		case "ports":
			for _, port := range body.Body {
				if len(port.Args) == 0 {
					c.Ports = append(c.Ports, portParse("", port.Name))
				} else {
					c.Ports = append(c.Ports, portParse(port.Name, port.Args[0]))
				}
			}
		case "env":
			asserts.ArgsMin(body, 2)
			c.Env = append(c.Env, envParse(body.Args[0], body.Args[1:]))
		case "envs":
			asserts.ArgsLen(body, 0)
			for _, env := range body.Body {
				c.Env = append(c.Env, envParse(env.Name, env.Args))
			}
		case "envFrom":
			asserts.ArgsMin(body, 1)
			fromType, name := utils.Split2(body.Args[0], ":")
			env := v1.EnvFromSource{
				Prefix: utils.Index(body.Args, 1),
			}
			if fromType == "secret" {
				env.ConfigMapRef = &v1.ConfigMapEnvSource{
					LocalObjectReference: v1.LocalObjectReference{Name: name},
				}
			} else {
				env.SecretRef = &v1.SecretEnvSource{
					LocalObjectReference: v1.LocalObjectReference{Name: name},
				}
			}
			c.EnvFrom = append(c.EnvFrom, env)

		case "device":
			asserts.ArgsLen(body, 2)
			c.VolumeDevices = append(c.VolumeDevices, v1.VolumeDevice{
				Name: body.Args[0], DevicePath: body.Args[1],
			})

		case "mount", "volumeMount":
			mountParse(body.Args, body, pod, &c)
		case "mounts", "volumeMounts":
			for _, directive := range body.Body {
				args := append([]string{directive.Name}, directive.Args...)
				mountParse(args, directive, pod, &c)
			}
		}
	}

	return c
}

func ContainerParse(d *config.Directive, spec *v1.PodSpec) {
	c := container(d, spec)
	spec.Containers = append(spec.Containers, c)
}

func InitContainerParse(d *config.Directive, spec *v1.PodSpec) {
	c := container(d, spec)
	spec.InitContainers = append(spec.InitContainers, c)
}
