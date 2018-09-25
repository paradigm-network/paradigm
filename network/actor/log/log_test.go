/****************************************************
Copyright 2018 The ont-eventbus Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*****************************************************/

/***************************************************
Copyright 2016 https://github.com/AsynkronIT/protoactor-go

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*****************************************************/

package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger_With(t *testing.T) {
	base := New(DebugLevel, "", Field{key: "first"})
	l := base.With(Field{key: "second"})

	assert.Equal(t, []Field{{key: "first"}, {key: "second"}}, l.context)
}

func Benchmark_OffLevel_TwoFields(b *testing.B) {
	l := New(MinLevel, "")
	for i := 0; i < b.N; i++ {
		l.Debug("foo", Int("bar", 32), Bool("fum", false))
	}
}

func Benchmark_OffLevel_OnlyContext(b *testing.B) {
	l := New(MinLevel, "", Int("bar", 32), Bool("fum", false))
	for i := 0; i < b.N; i++ {
		l.Debug("foo")
	}
}

func Benchmark_DebugLevel_OnlyContext_OneSubscriber(b *testing.B) {
	Unsubscribe(sub)
	s1 := Subscribe(func(Event) {})

	l := New(DebugLevel, "", Int("bar", 32), Bool("fum", false))
	for i := 0; i < b.N; i++ {
		l.Debug("foo")
	}
	Unsubscribe(s1)
}

func Benchmark_DebugLevel_OnlyContext_MultipleSubscribers(b *testing.B) {
	Unsubscribe(sub)
	s1 := Subscribe(func(Event) {})
	s2 := Subscribe(func(Event) {})

	l := New(DebugLevel, "", Int("bar", 32), Bool("fum", false))
	for i := 0; i < b.N; i++ {
		l.Debug("foo")
	}

	Unsubscribe(s1)
	Unsubscribe(s2)
}
