/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/validation"
)

// =============================================== Args related functions ==============================================

// ArgValidationConjunction returns a PositionalArgs validator that checks all provided validators in turn (all must pass).
func ArgValidationConjunction(validators ...cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, v := range validators {
			err := v(cmd, args)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// PositionalArg is a function for validating a single argument
type PositionalArg func(cmd *cobra.Command, arg string) error

// AtPosition returns a PositionalArgs that applies the single valued validator to the i-th argument.
// The actual number of arguments is not checked by this function (use cobra's MinimumNArgs, ExactArgs, etc)
func AtPosition(i int, validator PositionalArg) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		return validator(cmd, args[i])
	}
}

// KubernetesValidation turns a kubernetes-style validation function into a PositionalArg
func KubernetesValidation(k8s func(string) []string) PositionalArg {
	return func(cmd *cobra.Command, arg string) error {
		msgs := k8s(arg)
		if len(msgs) > 0 {
			return fmt.Errorf("%s", strings.Join(msgs, ", "))
		} else {
			return nil
		}
	}
}

func ValidName() PositionalArg {
	return KubernetesValidation(validation.IsDNS1123Subdomain)
}

// =============================================== Flags related functions =============================================

type FlagsValidator func(cmd *cobra.Command) error

// CobraEFunction is the type of functions cobra expects for Run, PreRun, etc that can return an error.
type CobraEFunction func(cmd *cobra.Command, args []string) error

// FlagsValidatorAsCobraRunE allows a FlagsValidator to be used as a CobraEFunction (typically PreRunE())
func FlagsValidatorAsCobraRunE(validator FlagsValidator) CobraEFunction {
	return func(cmd *cobra.Command, args []string) error {
		return validator(cmd)
	}
}

// FlagsValidationConjunction returns a FlagsValidator validator that checks all provided validators in turn (all must pass).
func FlagsValidationConjunction(validators ...FlagsValidator) FlagsValidator {
	return func(cmd *cobra.Command) error {
		for _, v := range validators {
			err := v(cmd)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// FlagsDependency returns a validator that will evaluate the given delegate if the provided flag has been set.
// Use to enforce scenarios such as "if --foo is set, then --bar must be set as well".
func FlagsDependency(flag string, delegate FlagsValidator) FlagsValidator {
	return func(cmd *cobra.Command) error {
		f := cmd.Flag(flag)
		if f == nil {
			panic(fmt.Sprintf("Expected to find flag named %q in command %q", flag, cmd.Use))
		}
		if f.Changed {
			// Flag set. Delegate condition must HOLD
			return delegate(cmd)
		} else {
			// Flag not set. Don't check delegate.
			return nil
		}
	}
}

// AtLeastOneOf returns a FlagsValidator that asserts that at least one of the passed in flags is set.
func AtLeastOneOf(flagNames ...string) FlagsValidator {
	return func(cmd *cobra.Command) error {
		for _, f := range flagNames {
			flag := cmd.Flag(f)
			if flag == nil {
				panic(fmt.Sprintf("Expected to find flag named %q in command %q", f, cmd.Use))
			}
			if flag.Changed {
				return nil
			}
		}
		return fmt.Errorf("at least one of --%s must be set", strings.Join(flagNames, ", --"))
	}
}

// AtMostOneOf returns a FlagsValidator that asserts that at most one of the passed in flags is set.
func AtMostOneOf(flagNames ...string) FlagsValidator {
	return func(cmd *cobra.Command) error {
		set := 0
		for _, f := range flagNames {
			flag := cmd.Flag(f)
			if flag == nil {
				panic(fmt.Sprintf("Expected to find flag named %q in command %q", f, cmd.Use))
			}
			if flag.Changed {
				set++
			}
		}
		if set > 1 {
			return fmt.Errorf("at most one of --%s must be set", strings.Join(flagNames, ", --"))
		} else {
			return nil
		}
	}
}

type broadcastStringValue []*string

func (bsv broadcastStringValue) Set(v string) error {
	for _, p := range bsv {
		*p = v
	}
	return nil
}

func (bsv broadcastStringValue) String() string {
	return *bsv[0]
}

func (bsv broadcastStringValue) Type() string {
	return "string"
}

func BroadcastStringValue(value string, ptrs ...*string) pflag.Value {
	if len(ptrs) < 1 {
		panic("At least one string pointer must be provided")
	}
	for i, _ := range ptrs {
		*ptrs[i] = value
	}
	return broadcastStringValue(ptrs)
}
