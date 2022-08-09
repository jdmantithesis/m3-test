// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/m3db/m3/src/query/generated/proto/prompb/types.proto

// Copyright (c) 2022 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

/*
	Package prompb is a generated protocol buffer package.

	It is generated from these files:
		github.com/m3db/m3/src/query/generated/proto/prompb/types.proto
		github.com/m3db/m3/src/query/generated/proto/prompb/remote.proto

	It has these top-level messages:
		Sample
		TimeSeries
		Label
		Labels
		LabelMatcher
		WriteRequest
		ReadRequest
		ReadResponse
		Query
		QueryResult
*/
package prompb

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"

import binary "encoding/binary"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type MetricType int32

const (
	MetricType_UNKNOWN         MetricType = 0
	MetricType_COUNTER         MetricType = 1
	MetricType_GAUGE           MetricType = 2
	MetricType_HISTOGRAM       MetricType = 3
	MetricType_GAUGE_HISTOGRAM MetricType = 4
	MetricType_SUMMARY         MetricType = 5
	MetricType_INFO            MetricType = 6
	MetricType_STATESET        MetricType = 7
)

var MetricType_name = map[int32]string{
	0: "UNKNOWN",
	1: "COUNTER",
	2: "GAUGE",
	3: "HISTOGRAM",
	4: "GAUGE_HISTOGRAM",
	5: "SUMMARY",
	6: "INFO",
	7: "STATESET",
}
var MetricType_value = map[string]int32{
	"UNKNOWN":         0,
	"COUNTER":         1,
	"GAUGE":           2,
	"HISTOGRAM":       3,
	"GAUGE_HISTOGRAM": 4,
	"SUMMARY":         5,
	"INFO":            6,
	"STATESET":        7,
}

func (x MetricType) String() string {
	return proto.EnumName(MetricType_name, int32(x))
}
func (MetricType) EnumDescriptor() ([]byte, []int) { return fileDescriptorTypes, []int{0} }

type M3Type int32

const (
	M3Type_M3_GAUGE   M3Type = 0
	M3Type_M3_COUNTER M3Type = 1
	M3Type_M3_TIMER   M3Type = 2
)

var M3Type_name = map[int32]string{
	0: "M3_GAUGE",
	1: "M3_COUNTER",
	2: "M3_TIMER",
}
var M3Type_value = map[string]int32{
	"M3_GAUGE":   0,
	"M3_COUNTER": 1,
	"M3_TIMER":   2,
}

func (x M3Type) String() string {
	return proto.EnumName(M3Type_name, int32(x))
}
func (M3Type) EnumDescriptor() ([]byte, []int) { return fileDescriptorTypes, []int{1} }

type Source int32

const (
	Source_PROMETHEUS   Source = 0
	Source_GRAPHITE     Source = 1
	Source_OPEN_METRICS Source = 2
)

var Source_name = map[int32]string{
	0: "PROMETHEUS",
	1: "GRAPHITE",
	2: "OPEN_METRICS",
}
var Source_value = map[string]int32{
	"PROMETHEUS":   0,
	"GRAPHITE":     1,
	"OPEN_METRICS": 2,
}

func (x Source) String() string {
	return proto.EnumName(Source_name, int32(x))
}
func (Source) EnumDescriptor() ([]byte, []int) { return fileDescriptorTypes, []int{2} }

type LabelMatcher_Type int32

const (
	LabelMatcher_EQ  LabelMatcher_Type = 0
	LabelMatcher_NEQ LabelMatcher_Type = 1
	LabelMatcher_RE  LabelMatcher_Type = 2
	LabelMatcher_NRE LabelMatcher_Type = 3
)

var LabelMatcher_Type_name = map[int32]string{
	0: "EQ",
	1: "NEQ",
	2: "RE",
	3: "NRE",
}
var LabelMatcher_Type_value = map[string]int32{
	"EQ":  0,
	"NEQ": 1,
	"RE":  2,
	"NRE": 3,
}

func (x LabelMatcher_Type) String() string {
	return proto.EnumName(LabelMatcher_Type_name, int32(x))
}
func (LabelMatcher_Type) EnumDescriptor() ([]byte, []int) { return fileDescriptorTypes, []int{4, 0} }

type Sample struct {
	Value     float64 `protobuf:"fixed64,1,opt,name=value,proto3" json:"value,omitempty"`
	Timestamp int64   `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (m *Sample) Reset()                    { *m = Sample{} }
func (m *Sample) String() string            { return proto.CompactTextString(m) }
func (*Sample) ProtoMessage()               {}
func (*Sample) Descriptor() ([]byte, []int) { return fileDescriptorTypes, []int{0} }

func (m *Sample) GetValue() float64 {
	if m != nil {
		return m.Value
	}
	return 0
}

func (m *Sample) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

type TimeSeries struct {
	Labels  []Label    `protobuf:"bytes,1,rep,name=labels" json:"labels"`
	Samples []Sample   `protobuf:"bytes,2,rep,name=samples" json:"samples"`
	Type    MetricType `protobuf:"varint,3,opt,name=type,proto3,enum=m3prometheus.MetricType" json:"type,omitempty"`
	Unit    string     `protobuf:"bytes,4,opt,name=unit,proto3" json:"unit,omitempty"`
	Help    string     `protobuf:"bytes,5,opt,name=help,proto3" json:"help,omitempty"`
	// NB: These are custom fields that M3 uses. They start at 101 so that they
	// should never clash with prometheus fields.
	M3Type M3Type `protobuf:"varint,101,opt,name=m3_type,json=m3Type,proto3,enum=m3prometheus.M3Type" json:"m3_type,omitempty"`
	Source Source `protobuf:"varint,102,opt,name=source,proto3,enum=m3prometheus.Source" json:"source,omitempty"`
}

func (m *TimeSeries) Reset()                    { *m = TimeSeries{} }
func (m *TimeSeries) String() string            { return proto.CompactTextString(m) }
func (*TimeSeries) ProtoMessage()               {}
func (*TimeSeries) Descriptor() ([]byte, []int) { return fileDescriptorTypes, []int{1} }

func (m *TimeSeries) GetLabels() []Label {
	if m != nil {
		return m.Labels
	}
	return nil
}

func (m *TimeSeries) GetSamples() []Sample {
	if m != nil {
		return m.Samples
	}
	return nil
}

func (m *TimeSeries) GetType() MetricType {
	if m != nil {
		return m.Type
	}
	return MetricType_UNKNOWN
}

func (m *TimeSeries) GetUnit() string {
	if m != nil {
		return m.Unit
	}
	return ""
}

func (m *TimeSeries) GetHelp() string {
	if m != nil {
		return m.Help
	}
	return ""
}

func (m *TimeSeries) GetM3Type() M3Type {
	if m != nil {
		return m.M3Type
	}
	return M3Type_M3_GAUGE
}

func (m *TimeSeries) GetSource() Source {
	if m != nil {
		return m.Source
	}
	return Source_PROMETHEUS
}

type Label struct {
	Name  []byte `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Value []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *Label) Reset()                    { *m = Label{} }
func (m *Label) String() string            { return proto.CompactTextString(m) }
func (*Label) ProtoMessage()               {}
func (*Label) Descriptor() ([]byte, []int) { return fileDescriptorTypes, []int{2} }

func (m *Label) GetName() []byte {
	if m != nil {
		return m.Name
	}
	return nil
}

func (m *Label) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type Labels struct {
	Labels []Label `protobuf:"bytes,1,rep,name=labels" json:"labels"`
}

func (m *Labels) Reset()                    { *m = Labels{} }
func (m *Labels) String() string            { return proto.CompactTextString(m) }
func (*Labels) ProtoMessage()               {}
func (*Labels) Descriptor() ([]byte, []int) { return fileDescriptorTypes, []int{3} }

func (m *Labels) GetLabels() []Label {
	if m != nil {
		return m.Labels
	}
	return nil
}

// Matcher specifies a rule, which can match or set of labels or not.
type LabelMatcher struct {
	Type  LabelMatcher_Type `protobuf:"varint,1,opt,name=type,proto3,enum=m3prometheus.LabelMatcher_Type" json:"type,omitempty"`
	Name  []byte            `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Value []byte            `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *LabelMatcher) Reset()                    { *m = LabelMatcher{} }
func (m *LabelMatcher) String() string            { return proto.CompactTextString(m) }
func (*LabelMatcher) ProtoMessage()               {}
func (*LabelMatcher) Descriptor() ([]byte, []int) { return fileDescriptorTypes, []int{4} }

func (m *LabelMatcher) GetType() LabelMatcher_Type {
	if m != nil {
		return m.Type
	}
	return LabelMatcher_EQ
}

func (m *LabelMatcher) GetName() []byte {
	if m != nil {
		return m.Name
	}
	return nil
}

func (m *LabelMatcher) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func init() {
	proto.RegisterType((*Sample)(nil), "m3prometheus.Sample")
	proto.RegisterType((*TimeSeries)(nil), "m3prometheus.TimeSeries")
	proto.RegisterType((*Label)(nil), "m3prometheus.Label")
	proto.RegisterType((*Labels)(nil), "m3prometheus.Labels")
	proto.RegisterType((*LabelMatcher)(nil), "m3prometheus.LabelMatcher")
	proto.RegisterEnum("m3prometheus.MetricType", MetricType_name, MetricType_value)
	proto.RegisterEnum("m3prometheus.M3Type", M3Type_name, M3Type_value)
	proto.RegisterEnum("m3prometheus.Source", Source_name, Source_value)
	proto.RegisterEnum("m3prometheus.LabelMatcher_Type", LabelMatcher_Type_name, LabelMatcher_Type_value)
}
func (m *Sample) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Sample) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Value != 0 {
		dAtA[i] = 0x9
		i++
		binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.Value))))
		i += 8
	}
	if m.Timestamp != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintTypes(dAtA, i, uint64(m.Timestamp))
	}
	return i, nil
}

func (m *TimeSeries) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TimeSeries) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Labels) > 0 {
		for _, msg := range m.Labels {
			dAtA[i] = 0xa
			i++
			i = encodeVarintTypes(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if len(m.Samples) > 0 {
		for _, msg := range m.Samples {
			dAtA[i] = 0x12
			i++
			i = encodeVarintTypes(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if m.Type != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintTypes(dAtA, i, uint64(m.Type))
	}
	if len(m.Unit) > 0 {
		dAtA[i] = 0x22
		i++
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Unit)))
		i += copy(dAtA[i:], m.Unit)
	}
	if len(m.Help) > 0 {
		dAtA[i] = 0x2a
		i++
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Help)))
		i += copy(dAtA[i:], m.Help)
	}
	if m.M3Type != 0 {
		dAtA[i] = 0xa8
		i++
		dAtA[i] = 0x6
		i++
		i = encodeVarintTypes(dAtA, i, uint64(m.M3Type))
	}
	if m.Source != 0 {
		dAtA[i] = 0xb0
		i++
		dAtA[i] = 0x6
		i++
		i = encodeVarintTypes(dAtA, i, uint64(m.Source))
	}
	return i, nil
}

func (m *Label) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Label) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	if len(m.Value) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Value)))
		i += copy(dAtA[i:], m.Value)
	}
	return i, nil
}

func (m *Labels) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Labels) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Labels) > 0 {
		for _, msg := range m.Labels {
			dAtA[i] = 0xa
			i++
			i = encodeVarintTypes(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *LabelMatcher) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *LabelMatcher) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Type != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintTypes(dAtA, i, uint64(m.Type))
	}
	if len(m.Name) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	if len(m.Value) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Value)))
		i += copy(dAtA[i:], m.Value)
	}
	return i, nil
}

func encodeVarintTypes(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Sample) Size() (n int) {
	var l int
	_ = l
	if m.Value != 0 {
		n += 9
	}
	if m.Timestamp != 0 {
		n += 1 + sovTypes(uint64(m.Timestamp))
	}
	return n
}

func (m *TimeSeries) Size() (n int) {
	var l int
	_ = l
	if len(m.Labels) > 0 {
		for _, e := range m.Labels {
			l = e.Size()
			n += 1 + l + sovTypes(uint64(l))
		}
	}
	if len(m.Samples) > 0 {
		for _, e := range m.Samples {
			l = e.Size()
			n += 1 + l + sovTypes(uint64(l))
		}
	}
	if m.Type != 0 {
		n += 1 + sovTypes(uint64(m.Type))
	}
	l = len(m.Unit)
	if l > 0 {
		n += 1 + l + sovTypes(uint64(l))
	}
	l = len(m.Help)
	if l > 0 {
		n += 1 + l + sovTypes(uint64(l))
	}
	if m.M3Type != 0 {
		n += 2 + sovTypes(uint64(m.M3Type))
	}
	if m.Source != 0 {
		n += 2 + sovTypes(uint64(m.Source))
	}
	return n
}

func (m *Label) Size() (n int) {
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovTypes(uint64(l))
	}
	l = len(m.Value)
	if l > 0 {
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}

func (m *Labels) Size() (n int) {
	var l int
	_ = l
	if len(m.Labels) > 0 {
		for _, e := range m.Labels {
			l = e.Size()
			n += 1 + l + sovTypes(uint64(l))
		}
	}
	return n
}

func (m *LabelMatcher) Size() (n int) {
	var l int
	_ = l
	if m.Type != 0 {
		n += 1 + sovTypes(uint64(m.Type))
	}
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovTypes(uint64(l))
	}
	l = len(m.Value)
	if l > 0 {
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}

func sovTypes(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozTypes(x uint64) (n int) {
	return sovTypes(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Sample) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Sample: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Sample: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.Value = float64(math.Float64frombits(v))
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Timestamp", wireType)
			}
			m.Timestamp = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Timestamp |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *TimeSeries) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: TimeSeries: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TimeSeries: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Labels", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Labels = append(m.Labels, Label{})
			if err := m.Labels[len(m.Labels)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Samples", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Samples = append(m.Samples, Sample{})
			if err := m.Samples[len(m.Samples)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= (MetricType(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Unit", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Unit = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Help", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Help = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 101:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field M3Type", wireType)
			}
			m.M3Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.M3Type |= (M3Type(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 102:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Source", wireType)
			}
			m.Source = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Source |= (Source(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Label) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Label: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Label: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = append(m.Name[:0], dAtA[iNdEx:postIndex]...)
			if m.Name == nil {
				m.Name = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Value = append(m.Value[:0], dAtA[iNdEx:postIndex]...)
			if m.Value == nil {
				m.Value = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Labels) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Labels: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Labels: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Labels", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Labels = append(m.Labels, Label{})
			if err := m.Labels[len(m.Labels)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *LabelMatcher) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: LabelMatcher: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: LabelMatcher: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= (LabelMatcher_Type(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = append(m.Name[:0], dAtA[iNdEx:postIndex]...)
			if m.Name == nil {
				m.Name = []byte{}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Value = append(m.Value[:0], dAtA[iNdEx:postIndex]...)
			if m.Value == nil {
				m.Value = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTypes(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthTypes
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowTypes
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipTypes(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthTypes = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTypes   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/m3db/m3/src/query/generated/proto/prompb/types.proto", fileDescriptorTypes)
}

var fileDescriptorTypes = []byte{
	// 605 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x53, 0xcd, 0x6a, 0xdb, 0x4c,
	0x14, 0xf5, 0x48, 0xb2, 0x1c, 0xdf, 0xf8, 0xcb, 0x37, 0x4c, 0xb2, 0x10, 0xa5, 0x38, 0xc6, 0x2b,
	0x13, 0x12, 0x8b, 0x44, 0x59, 0x14, 0x5a, 0x28, 0x4e, 0x98, 0x3a, 0xa6, 0x91, 0x9c, 0x8c, 0x64,
	0x4a, 0xbb, 0x31, 0xb2, 0x33, 0xb1, 0x0d, 0x9e, 0x48, 0xd5, 0x4f, 0x21, 0x7d, 0x8a, 0xee, 0xfa,
	0x0a, 0x7d, 0x94, 0x2c, 0xfb, 0x04, 0xa5, 0xa4, 0x2f, 0x52, 0x66, 0xc6, 0xc5, 0x49, 0xc9, 0xa6,
	0x1b, 0x7b, 0xee, 0xb9, 0xe7, 0xdc, 0x7b, 0x74, 0x86, 0x81, 0xd7, 0xb3, 0x45, 0x31, 0x2f, 0x27,
	0xdd, 0x69, 0x22, 0x5c, 0xe1, 0x5d, 0x4d, 0x5c, 0xe1, 0xb9, 0x79, 0x36, 0x75, 0x3f, 0x96, 0x3c,
	0xbb, 0x75, 0x67, 0xfc, 0x86, 0x67, 0x71, 0xc1, 0xaf, 0xdc, 0x34, 0x4b, 0x8a, 0x44, 0xfe, 0x8a,
	0x74, 0xe2, 0x16, 0xb7, 0x29, 0xcf, 0xbb, 0x0a, 0x22, 0x0d, 0xe1, 0x49, 0x94, 0x17, 0x73, 0x5e,
	0xe6, 0xcf, 0x0e, 0x1e, 0x8c, 0x9b, 0x25, 0xb3, 0x44, 0xeb, 0x26, 0xe5, 0xb5, 0xaa, 0xf4, 0x10,
	0x79, 0xd2, 0xe2, 0xf6, 0x2b, 0xb0, 0xc3, 0x58, 0xa4, 0x4b, 0x4e, 0x76, 0xa0, 0xfa, 0x29, 0x5e,
	0x96, 0xdc, 0x41, 0x2d, 0xd4, 0x41, 0x4c, 0x17, 0xe4, 0x39, 0xd4, 0x8b, 0x85, 0xe0, 0x79, 0x11,
	0x8b, 0xd4, 0x31, 0x5a, 0xa8, 0x63, 0xb2, 0x35, 0xd0, 0xfe, 0x66, 0x00, 0x44, 0x0b, 0xc1, 0x43,
	0x9e, 0x2d, 0x78, 0x4e, 0x0e, 0xc1, 0x5e, 0xc6, 0x13, 0xbe, 0xcc, 0x1d, 0xd4, 0x32, 0x3b, 0x9b,
	0x47, 0xdb, 0xdd, 0x87, 0xd6, 0xba, 0xe7, 0xb2, 0x77, 0x62, 0xdd, 0xfd, 0xd8, 0xad, 0xb0, 0x15,
	0x91, 0x1c, 0x43, 0x2d, 0x57, 0xfb, 0x73, 0xc7, 0x50, 0x9a, 0x9d, 0xc7, 0x1a, 0x6d, 0x6e, 0x25,
	0xfa, 0x43, 0x25, 0xfb, 0x60, 0xc9, 0x04, 0x1c, 0xb3, 0x85, 0x3a, 0x5b, 0x47, 0xce, 0x63, 0x89,
	0xcf, 0x8b, 0x6c, 0x31, 0x8d, 0x6e, 0x53, 0xce, 0x14, 0x8b, 0x10, 0xb0, 0xca, 0x9b, 0x45, 0xe1,
	0x58, 0x2d, 0xd4, 0xa9, 0x33, 0x75, 0x96, 0xd8, 0x9c, 0x2f, 0x53, 0xa7, 0xaa, 0x31, 0x79, 0x26,
	0x07, 0x50, 0x13, 0xde, 0x58, 0x0d, 0xe6, 0x6a, 0xf0, 0x5f, 0x5e, 0x7c, 0x4f, 0x0d, 0xb5, 0x85,
	0xfa, 0x27, 0xfb, 0x60, 0xe7, 0x49, 0x99, 0x4d, 0xb9, 0x73, 0xfd, 0x14, 0x3b, 0x54, 0x3d, 0xb6,
	0xe2, 0xb4, 0x0f, 0xa1, 0xaa, 0xbe, 0x5f, 0x6e, 0xbe, 0x89, 0x85, 0x8e, 0xb9, 0xc1, 0xd4, 0x79,
	0x9d, 0xbd, 0xa1, 0x40, 0x5d, 0xb4, 0x5f, 0x82, 0x7d, 0xae, 0x53, 0xfa, 0xf7, 0x60, 0xdb, 0x5f,
	0x11, 0x34, 0x14, 0xee, 0xc7, 0xc5, 0x74, 0xce, 0x33, 0xe2, 0xad, 0x32, 0x43, 0xca, 0xec, 0xee,
	0x13, 0x13, 0x56, 0xcc, 0xee, 0xe3, 0xe8, 0x94, 0x59, 0xe3, 0x29, 0xb3, 0xe6, 0x43, 0xb3, 0x1d,
	0xb0, 0x54, 0x2a, 0x36, 0x18, 0xf4, 0x12, 0x57, 0x48, 0x0d, 0xcc, 0x80, 0x5e, 0x62, 0x24, 0x01,
	0x46, 0xb1, 0xa1, 0x00, 0x46, 0xb1, 0xb9, 0xf7, 0x19, 0x60, 0x7d, 0x45, 0x64, 0x13, 0x6a, 0xa3,
	0xe0, 0x6d, 0x30, 0x7c, 0x17, 0xe0, 0x8a, 0x2c, 0x4e, 0x87, 0xa3, 0x20, 0xa2, 0x0c, 0x23, 0x52,
	0x87, 0x6a, 0xbf, 0x37, 0xea, 0x4b, 0xed, 0x7f, 0x50, 0x3f, 0x1b, 0x84, 0xd1, 0xb0, 0xcf, 0x7a,
	0x3e, 0x36, 0xc9, 0x36, 0xfc, 0xaf, 0x3a, 0xe3, 0x35, 0x68, 0x49, 0x6d, 0x38, 0xf2, 0xfd, 0x1e,
	0x7b, 0x8f, 0xab, 0x64, 0x03, 0xac, 0x41, 0xf0, 0x66, 0x88, 0x6d, 0xd2, 0x80, 0x8d, 0x30, 0xea,
	0x45, 0x34, 0xa4, 0x11, 0xae, 0xed, 0x1d, 0x83, 0xad, 0x6f, 0x51, 0xe2, 0xbe, 0x37, 0xd6, 0x0b,
	0x2a, 0x64, 0x0b, 0xc0, 0xf7, 0xc6, 0xeb, 0xdd, 0xba, 0x1b, 0x0d, 0x7c, 0xca, 0xb0, 0xb1, 0xf7,
	0x02, 0x6c, 0x7d, 0x9b, 0x92, 0x77, 0xc1, 0x86, 0x3e, 0x8d, 0xce, 0xe8, 0x28, 0xc4, 0x15, 0xc9,
	0xeb, 0xb3, 0xde, 0xc5, 0xd9, 0x20, 0xa2, 0x18, 0x11, 0x0c, 0x8d, 0xe1, 0x05, 0x0d, 0xc6, 0x3e,
	0x8d, 0xd8, 0xe0, 0x34, 0xc4, 0xc6, 0x89, 0x73, 0x77, 0xdf, 0x44, 0xdf, 0xef, 0x9b, 0xe8, 0xe7,
	0x7d, 0x13, 0x7d, 0xf9, 0xd5, 0xac, 0x7c, 0xb0, 0xf5, 0x0b, 0x9e, 0xd8, 0xea, 0xfd, 0x79, 0xbf,
	0x03, 0x00, 0x00, 0xff, 0xff, 0x3e, 0xb3, 0xe8, 0x2d, 0xff, 0x03, 0x00, 0x00,
}
