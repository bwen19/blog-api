// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.20.1
// source: notification_message.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Notification struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Kind     string                 `protobuf:"bytes,2,opt,name=kind,proto3" json:"kind,omitempty"`
	Title    string                 `protobuf:"bytes,3,opt,name=title,proto3" json:"title,omitempty"`
	Content  string                 `protobuf:"bytes,4,opt,name=content,proto3" json:"content,omitempty"`
	Unread   bool                   `protobuf:"varint,5,opt,name=unread,proto3" json:"unread,omitempty"`
	CreateAt *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=create_at,json=createAt,proto3" json:"create_at,omitempty"`
}

func (x *Notification) Reset() {
	*x = Notification{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notification_message_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Notification) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Notification) ProtoMessage() {}

func (x *Notification) ProtoReflect() protoreflect.Message {
	mi := &file_notification_message_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Notification.ProtoReflect.Descriptor instead.
func (*Notification) Descriptor() ([]byte, []int) {
	return file_notification_message_proto_rawDescGZIP(), []int{0}
}

func (x *Notification) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Notification) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *Notification) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Notification) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

func (x *Notification) GetUnread() bool {
	if x != nil {
		return x.Unread
	}
	return false
}

func (x *Notification) GetCreateAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreateAt
	}
	return nil
}

type LeaveMessageRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Title   string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Content string `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *LeaveMessageRequest) Reset() {
	*x = LeaveMessageRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notification_message_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LeaveMessageRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LeaveMessageRequest) ProtoMessage() {}

func (x *LeaveMessageRequest) ProtoReflect() protoreflect.Message {
	mi := &file_notification_message_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LeaveMessageRequest.ProtoReflect.Descriptor instead.
func (*LeaveMessageRequest) Descriptor() ([]byte, []int) {
	return file_notification_message_proto_rawDescGZIP(), []int{1}
}

func (x *LeaveMessageRequest) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *LeaveMessageRequest) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

type DeleteNotifsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NotificationIds []int64 `protobuf:"varint,1,rep,packed,name=notification_ids,json=notificationIds,proto3" json:"notification_ids,omitempty"`
}

func (x *DeleteNotifsRequest) Reset() {
	*x = DeleteNotifsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notification_message_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteNotifsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteNotifsRequest) ProtoMessage() {}

func (x *DeleteNotifsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_notification_message_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteNotifsRequest.ProtoReflect.Descriptor instead.
func (*DeleteNotifsRequest) Descriptor() ([]byte, []int) {
	return file_notification_message_proto_rawDescGZIP(), []int{2}
}

func (x *DeleteNotifsRequest) GetNotificationIds() []int64 {
	if x != nil {
		return x.NotificationIds
	}
	return nil
}

type ListNotifsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PageId   int32  `protobuf:"varint,1,opt,name=page_id,json=pageId,proto3" json:"page_id,omitempty"`
	PageSize int32  `protobuf:"varint,2,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	Kind     string `protobuf:"bytes,3,opt,name=kind,proto3" json:"kind,omitempty"`
}

func (x *ListNotifsRequest) Reset() {
	*x = ListNotifsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notification_message_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListNotifsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListNotifsRequest) ProtoMessage() {}

func (x *ListNotifsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_notification_message_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListNotifsRequest.ProtoReflect.Descriptor instead.
func (*ListNotifsRequest) Descriptor() ([]byte, []int) {
	return file_notification_message_proto_rawDescGZIP(), []int{3}
}

func (x *ListNotifsRequest) GetPageId() int32 {
	if x != nil {
		return x.PageId
	}
	return 0
}

func (x *ListNotifsRequest) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

func (x *ListNotifsRequest) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

type ListNotifsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Total         int64           `protobuf:"varint,1,opt,name=total,proto3" json:"total,omitempty"`
	UnreadCount   int64           `protobuf:"varint,2,opt,name=unread_count,json=unreadCount,proto3" json:"unread_count,omitempty"`
	Notifications []*Notification `protobuf:"bytes,3,rep,name=notifications,proto3" json:"notifications,omitempty"`
}

func (x *ListNotifsResponse) Reset() {
	*x = ListNotifsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notification_message_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListNotifsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListNotifsResponse) ProtoMessage() {}

func (x *ListNotifsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_notification_message_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListNotifsResponse.ProtoReflect.Descriptor instead.
func (*ListNotifsResponse) Descriptor() ([]byte, []int) {
	return file_notification_message_proto_rawDescGZIP(), []int{4}
}

func (x *ListNotifsResponse) GetTotal() int64 {
	if x != nil {
		return x.Total
	}
	return 0
}

func (x *ListNotifsResponse) GetUnreadCount() int64 {
	if x != nil {
		return x.UnreadCount
	}
	return 0
}

func (x *ListNotifsResponse) GetNotifications() []*Notification {
	if x != nil {
		return x.Notifications
	}
	return nil
}

type ListMessagesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PageId   int32 `protobuf:"varint,1,opt,name=page_id,json=pageId,proto3" json:"page_id,omitempty"`
	PageSize int32 `protobuf:"varint,2,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
}

func (x *ListMessagesRequest) Reset() {
	*x = ListMessagesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notification_message_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListMessagesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListMessagesRequest) ProtoMessage() {}

func (x *ListMessagesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_notification_message_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListMessagesRequest.ProtoReflect.Descriptor instead.
func (*ListMessagesRequest) Descriptor() ([]byte, []int) {
	return file_notification_message_proto_rawDescGZIP(), []int{5}
}

func (x *ListMessagesRequest) GetPageId() int32 {
	if x != nil {
		return x.PageId
	}
	return 0
}

func (x *ListMessagesRequest) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

type ListMessagesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Total    int64                               `protobuf:"varint,1,opt,name=total,proto3" json:"total,omitempty"`
	Messages []*ListMessagesResponse_MessageItem `protobuf:"bytes,2,rep,name=messages,proto3" json:"messages,omitempty"`
}

func (x *ListMessagesResponse) Reset() {
	*x = ListMessagesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notification_message_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListMessagesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListMessagesResponse) ProtoMessage() {}

func (x *ListMessagesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_notification_message_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListMessagesResponse.ProtoReflect.Descriptor instead.
func (*ListMessagesResponse) Descriptor() ([]byte, []int) {
	return file_notification_message_proto_rawDescGZIP(), []int{6}
}

func (x *ListMessagesResponse) GetTotal() int64 {
	if x != nil {
		return x.Total
	}
	return 0
}

func (x *ListMessagesResponse) GetMessages() []*ListMessagesResponse_MessageItem {
	if x != nil {
		return x.Messages
	}
	return nil
}

type CheckMessagesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MessageIds []int64 `protobuf:"varint,1,rep,packed,name=message_ids,json=messageIds,proto3" json:"message_ids,omitempty"`
}

func (x *CheckMessagesRequest) Reset() {
	*x = CheckMessagesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notification_message_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckMessagesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckMessagesRequest) ProtoMessage() {}

func (x *CheckMessagesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_notification_message_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckMessagesRequest.ProtoReflect.Descriptor instead.
func (*CheckMessagesRequest) Descriptor() ([]byte, []int) {
	return file_notification_message_proto_rawDescGZIP(), []int{7}
}

func (x *CheckMessagesRequest) GetMessageIds() []int64 {
	if x != nil {
		return x.MessageIds
	}
	return nil
}

type ListMessagesResponse_MessageItem struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Kind     string                 `protobuf:"bytes,2,opt,name=kind,proto3" json:"kind,omitempty"`
	Title    string                 `protobuf:"bytes,3,opt,name=title,proto3" json:"title,omitempty"`
	User     *User                  `protobuf:"bytes,4,opt,name=user,proto3" json:"user,omitempty"`
	Content  string                 `protobuf:"bytes,5,opt,name=content,proto3" json:"content,omitempty"`
	Unread   bool                   `protobuf:"varint,6,opt,name=unread,proto3" json:"unread,omitempty"`
	CreateAt *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=create_at,json=createAt,proto3" json:"create_at,omitempty"`
}

func (x *ListMessagesResponse_MessageItem) Reset() {
	*x = ListMessagesResponse_MessageItem{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notification_message_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListMessagesResponse_MessageItem) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListMessagesResponse_MessageItem) ProtoMessage() {}

func (x *ListMessagesResponse_MessageItem) ProtoReflect() protoreflect.Message {
	mi := &file_notification_message_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListMessagesResponse_MessageItem.ProtoReflect.Descriptor instead.
func (*ListMessagesResponse_MessageItem) Descriptor() ([]byte, []int) {
	return file_notification_message_proto_rawDescGZIP(), []int{6, 0}
}

func (x *ListMessagesResponse_MessageItem) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *ListMessagesResponse_MessageItem) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *ListMessagesResponse_MessageItem) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *ListMessagesResponse_MessageItem) GetUser() *User {
	if x != nil {
		return x.User
	}
	return nil
}

func (x *ListMessagesResponse_MessageItem) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

func (x *ListMessagesResponse_MessageItem) GetUnread() bool {
	if x != nil {
		return x.Unread
	}
	return false
}

func (x *ListMessagesResponse_MessageItem) GetCreateAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreateAt
	}
	return nil
}

var File_notification_message_proto protoreflect.FileDescriptor

var file_notification_message_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70, 0x62,
	0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x14, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xb3, 0x01, 0x0a, 0x0c, 0x4e, 0x6f, 0x74, 0x69,
	0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x14, 0x0a, 0x05,
	0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x69, 0x74,
	0x6c, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x16, 0x0a, 0x06,
	0x75, 0x6e, 0x72, 0x65, 0x61, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x75, 0x6e,
	0x72, 0x65, 0x61, 0x64, 0x12, 0x37, 0x0a, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x5f, 0x61,
	0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x52, 0x08, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x74, 0x22, 0x45, 0x0a,
	0x13, 0x4c, 0x65, 0x61, 0x76, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x22, 0x40, 0x0a, 0x13, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x4e, 0x6f,
	0x74, 0x69, 0x66, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x29, 0x0a, 0x10, 0x6e,
	0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x03, 0x52, 0x0f, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x49, 0x64, 0x73, 0x22, 0x5d, 0x0a, 0x11, 0x4c, 0x69, 0x73, 0x74, 0x4e, 0x6f,
	0x74, 0x69, 0x66, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x70,
	0x61, 0x67, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x70, 0x61,
	0x67, 0x65, 0x49, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6b, 0x69, 0x6e, 0x64, 0x22, 0x85, 0x01, 0x0a, 0x12, 0x4c, 0x69, 0x73, 0x74, 0x4e, 0x6f,
	0x74, 0x69, 0x66, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05,
	0x74, 0x6f, 0x74, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x74, 0x6f, 0x74,
	0x61, 0x6c, 0x12, 0x21, 0x0a, 0x0c, 0x75, 0x6e, 0x72, 0x65, 0x61, 0x64, 0x5f, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x75, 0x6e, 0x72, 0x65, 0x61, 0x64,
	0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x36, 0x0a, 0x0d, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x70,
	0x62, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0d,
	0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x4b, 0x0a,
	0x13, 0x4c, 0x69, 0x73, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x70, 0x61, 0x67, 0x65, 0x49, 0x64, 0x12, 0x1b, 0x0a,
	0x09, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x22, 0xc1, 0x02, 0x0a, 0x14, 0x4c,
	0x69, 0x73, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x12, 0x40, 0x0a, 0x08, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x70, 0x62,
	0x2e, 0x4c, 0x69, 0x73, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x74, 0x65,
	0x6d, 0x52, 0x08, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x1a, 0xd0, 0x01, 0x0a, 0x0b,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6b,
	0x69, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12,
	0x14, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x74, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x1c, 0x0a, 0x04, 0x75, 0x73, 0x65, 0x72, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x08, 0x2e, 0x70, 0x62, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x52, 0x04, 0x75,
	0x73, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x16, 0x0a,
	0x06, 0x75, 0x6e, 0x72, 0x65, 0x61, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x75,
	0x6e, 0x72, 0x65, 0x61, 0x64, 0x12, 0x37, 0x0a, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x5f,
	0x61, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x52, 0x08, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x74, 0x22, 0x37,
	0x0a, 0x14, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x03, 0x52, 0x0a, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x49, 0x64, 0x73, 0x42, 0x10, 0x5a, 0x0e, 0x62, 0x6c, 0x6f, 0x67, 0x2f,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_notification_message_proto_rawDescOnce sync.Once
	file_notification_message_proto_rawDescData = file_notification_message_proto_rawDesc
)

func file_notification_message_proto_rawDescGZIP() []byte {
	file_notification_message_proto_rawDescOnce.Do(func() {
		file_notification_message_proto_rawDescData = protoimpl.X.CompressGZIP(file_notification_message_proto_rawDescData)
	})
	return file_notification_message_proto_rawDescData
}

var file_notification_message_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_notification_message_proto_goTypes = []interface{}{
	(*Notification)(nil),                     // 0: pb.Notification
	(*LeaveMessageRequest)(nil),              // 1: pb.LeaveMessageRequest
	(*DeleteNotifsRequest)(nil),              // 2: pb.DeleteNotifsRequest
	(*ListNotifsRequest)(nil),                // 3: pb.ListNotifsRequest
	(*ListNotifsResponse)(nil),               // 4: pb.ListNotifsResponse
	(*ListMessagesRequest)(nil),              // 5: pb.ListMessagesRequest
	(*ListMessagesResponse)(nil),             // 6: pb.ListMessagesResponse
	(*CheckMessagesRequest)(nil),             // 7: pb.CheckMessagesRequest
	(*ListMessagesResponse_MessageItem)(nil), // 8: pb.ListMessagesResponse.MessageItem
	(*timestamppb.Timestamp)(nil),            // 9: google.protobuf.Timestamp
	(*User)(nil),                             // 10: pb.User
}
var file_notification_message_proto_depIdxs = []int32{
	9,  // 0: pb.Notification.create_at:type_name -> google.protobuf.Timestamp
	0,  // 1: pb.ListNotifsResponse.notifications:type_name -> pb.Notification
	8,  // 2: pb.ListMessagesResponse.messages:type_name -> pb.ListMessagesResponse.MessageItem
	10, // 3: pb.ListMessagesResponse.MessageItem.user:type_name -> pb.User
	9,  // 4: pb.ListMessagesResponse.MessageItem.create_at:type_name -> google.protobuf.Timestamp
	5,  // [5:5] is the sub-list for method output_type
	5,  // [5:5] is the sub-list for method input_type
	5,  // [5:5] is the sub-list for extension type_name
	5,  // [5:5] is the sub-list for extension extendee
	0,  // [0:5] is the sub-list for field type_name
}

func init() { file_notification_message_proto_init() }
func file_notification_message_proto_init() {
	if File_notification_message_proto != nil {
		return
	}
	file_common_message_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_notification_message_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Notification); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_notification_message_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LeaveMessageRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_notification_message_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteNotifsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_notification_message_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListNotifsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_notification_message_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListNotifsResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_notification_message_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListMessagesRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_notification_message_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListMessagesResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_notification_message_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckMessagesRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_notification_message_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListMessagesResponse_MessageItem); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_notification_message_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_notification_message_proto_goTypes,
		DependencyIndexes: file_notification_message_proto_depIdxs,
		MessageInfos:      file_notification_message_proto_msgTypes,
	}.Build()
	File_notification_message_proto = out.File
	file_notification_message_proto_rawDesc = nil
	file_notification_message_proto_goTypes = nil
	file_notification_message_proto_depIdxs = nil
}
