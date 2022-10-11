package api

import (
	"github.com/bwen19/blog/grpc/pb"
	"github.com/bwen19/blog/psql/db"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertUser(user db.User) *pb.User {
	return &pb.User{
		Id:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Intro:    user.Intro,
		Role:     user.Role,
	}
}

func convertListUsers(users []db.ListUsersRow) *pb.ListUsersResponse {
	if len(users) == 0 {
		return &pb.ListUsersResponse{}
	}

	rspUsers := make([]*pb.ListUsersResponse_User, 0, 5)
	for _, user := range users {
		pbUser := &pb.ListUsersResponse_User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Role:      user.Role,
			Deleted:   user.Deleted,
			PostCount: user.PostCount,
			CreateAt:  timestamppb.New(user.CreateAt),
		}
		rspUsers = append(rspUsers, pbUser)
	}

	return &pb.ListUsersResponse{
		Total: users[0].Total,
		Users: rspUsers,
	}
}

func convertListSessions(sessions []db.ListSessionsRow) *pb.ListSessionsResponse {
	if len(sessions) == 0 {
		return &pb.ListSessionsResponse{}
	}

	rspSessions := make([]*pb.ListSessionsResponse_SessionItem, 0, 5)
	for _, session := range sessions {
		pbSession := &pb.ListSessionsResponse_SessionItem{
			Id:        session.ID.String(),
			UserAgent: session.UserAgent,
			ClientIp:  session.ClientIp,
			CreateAt:  timestamppb.New(session.CreateAt),
			ExpiresAt: timestamppb.New(session.ExpiresAt),
		}
		rspSessions = append(rspSessions, pbSession)
	}

	return &pb.ListSessionsResponse{
		Total:    sessions[0].Total,
		Sessions: rspSessions,
	}
}

func convertListNotifs(notifs []db.ListNotificationsRow, nReads int64) *pb.ListNotifsResponse {
	if len(notifs) == 0 {
		return &pb.ListNotifsResponse{}
	}

	rspNotifs := make([]*pb.Notification, 0, 5)
	for _, notif := range notifs {
		pbNotif := &pb.Notification{
			Id:       notif.ID,
			Kind:     notif.Kind,
			Title:    notif.Title,
			Content:  notif.Content,
			Unread:   notif.Unread,
			CreateAt: timestamppb.New(notif.CreateAt),
		}
		rspNotifs = append(rspNotifs, pbNotif)
	}

	return &pb.ListNotifsResponse{
		Total:         notifs[0].Total,
		UnreadCount:   notifs[0].UnreadCount - nReads,
		SystemCount:   notifs[0].SystemCount,
		ReplyCount:    notifs[0].ReplyCount,
		Notifications: rspNotifs,
	}
}

func convertLisMessages(messages []db.ListMessagesRow) *pb.ListMessagesResponse {
	if len(messages) == 0 {
		return &pb.ListMessagesResponse{}
	}

	rspMessages := make([]*pb.ListMessagesResponse_MessageItem, 0, 5)
	for _, msg := range messages {
		user := &pb.User{
			Id:       msg.UserID,
			Username: msg.Username,
			Avatar:   msg.Avatar,
			Email:    msg.Email,
		}
		pbMsg := &pb.ListMessagesResponse_MessageItem{
			Id:       msg.ID,
			Kind:     msg.Kind,
			Title:    msg.Title,
			User:     user,
			Content:  msg.Content,
			Unread:   msg.Unread,
			CreateAt: timestamppb.New(msg.CreateAt),
		}
		rspMessages = append(rspMessages, pbMsg)
	}

	return &pb.ListMessagesResponse{
		Total:       messages[0].Total,
		UnreadCount: messages[0].UnreadCount,
		Messages:    rspMessages,
	}
}

func convertListFollowers(followers []db.ListFollowersRow) *pb.ListFollowsResponse {
	if len(followers) == 0 {
		return &pb.ListFollowsResponse{}
	}

	rspUsers := make([]*pb.UserInfo, 0, 5)
	for _, follower := range followers {
		pbUserInfo := &pb.UserInfo{
			Id:       follower.UserID,
			Username: follower.Username,
			Avatar:   follower.Avatar,
			Intro:    follower.Intro,
			Followed: follower.Followed.Valid,
		}
		rspUsers = append(rspUsers, pbUserInfo)
	}

	return &pb.ListFollowsResponse{
		Total: followers[0].Total,
		Users: rspUsers,
	}
}

func convertListFollowings(followings []db.ListFollowingsRow) *pb.ListFollowsResponse {
	if len(followings) == 0 {
		return &pb.ListFollowsResponse{}
	}

	rspUsers := make([]*pb.UserInfo, 0, 5)
	for _, follower := range followings {
		pbUserInfo := &pb.UserInfo{
			Id:       follower.UserID,
			Username: follower.Username,
			Avatar:   follower.Avatar,
			Intro:    follower.Intro,
			Followed: follower.Followed.Valid,
		}
		rspUsers = append(rspUsers, pbUserInfo)
	}

	return &pb.ListFollowsResponse{
		Total: followings[0].Total,
		Users: rspUsers,
	}
}

func convertCategory(category db.Category) *pb.Category {
	return &pb.Category{
		Id:   category.ID,
		Name: category.Name,
	}
}

func convertCategories(categories []db.Category) []*pb.Category {
	rspCategories := make([]*pb.Category, 0, 2)
	for _, category := range categories {
		rspCategories = append(rspCategories, convertCategory(category))
	}
	return rspCategories
}

func convertListCategories(categories []db.ListCategoriesRow) *pb.ListCategoriesResponse {
	rspCategories := []*pb.ListCategoriesResponse_CategoryItem{}
	for _, category := range categories {
		pbCategory := &pb.ListCategoriesResponse_CategoryItem{
			Id:        category.ID,
			Name:      category.Name,
			PostCount: category.PostCount,
		}
		rspCategories = append(rspCategories, pbCategory)
	}
	return &pb.ListCategoriesResponse{
		Categories: rspCategories,
	}
}

func convertTag(tag db.Tag) *pb.Tag {
	return &pb.Tag{
		Id:   tag.ID,
		Name: tag.Name,
	}
}

func convertTags(tags []db.Tag) []*pb.Tag {
	rspTags := make([]*pb.Tag, 0, 5)
	for _, tag := range tags {
		rspTags = append(rspTags, convertTag(tag))
	}
	return rspTags
}

func convertListTags(tags []db.ListTagsRow) *pb.ListTagsResponse {
	if len(tags) == 0 {
		return &pb.ListTagsResponse{}
	}

	rspTags := make([]*pb.ListTagsResponse_TagItem, 0, 5)
	for _, tag := range tags {
		pbTag := &pb.ListTagsResponse_TagItem{
			Id:        tag.ID,
			Name:      tag.Name,
			PostCount: tag.PostCount,
		}
		rspTags = append(rspTags, pbTag)
	}

	return &pb.ListTagsResponse{
		Total: tags[0].Total,
		Tags:  rspTags,
	}
}

func convertNewPost(post db.CreateNewPostRow) *pb.Post {
	return &pb.Post{
		Id:         post.ID,
		Title:      post.Title,
		CoverImage: post.CoverImage,
		Content:    post.Content,
		Featured:   post.Featured,
		Status:     post.Status,
	}
}

func convertUpdatePost(post db.Post, content db.PostContent, categories []db.Category, tags []db.Tag) *pb.UpdatePostResponse {
	pbPost := &pb.Post{
		Id:         post.ID,
		Title:      post.Title,
		CoverImage: post.CoverImage,
		Content:    content.Content,
		Categories: convertCategories(categories),
		Tags:       convertTags(tags),
		Featured:   post.Featured,
		Status:     post.Status,
	}
	return &pb.UpdatePostResponse{Post: pbPost}
}

func convertListPosts(posts []db.ListPostsRow) *pb.ListPostsResponse {
	if len(posts) == 0 {
		return &pb.ListPostsResponse{}
	}

	rspPosts := []*pb.ListPostsResponse_PostItem{}
	for _, post := range posts {
		author := &pb.UserItem{
			Id:       post.AuthorID,
			Username: post.Username,
			Avatar:   post.Avatar,
		}

		categories := make([]*pb.Category, 0, 2)
		for i, categoryID := range post.CategoryIds {
			category := &pb.Category{
				Id:   categoryID,
				Name: post.CategoryNames[i],
			}
			categories = append(categories, category)
		}

		tags := make([]*pb.Tag, 0, 5)
		for i, tagID := range post.TagIds {
			tag := &pb.Tag{
				Id:   tagID,
				Name: post.TagNames[i],
			}
			tags = append(tags, tag)
		}

		pbPost := &pb.ListPostsResponse_PostItem{
			Id:         post.ID,
			Title:      post.Title,
			Author:     author,
			Categories: categories,
			Tags:       tags,
			Status:     post.Status,
			Featured:   post.Featured,
			ViewCount:  post.ViewCount,
			UpdateAt:   timestamppb.New(post.UpdateAt),
			PublishAt:  timestamppb.New(post.PublishAt),
		}
		rspPosts = append(rspPosts, pbPost)
	}

	return &pb.ListPostsResponse{
		Total: posts[0].Total,
		Posts: rspPosts,
	}
}

func convertGetPost(post db.GetPostRow) *pb.GetPostResponse {
	categories := make([]*pb.Category, 0, 2)
	for i, categoryID := range post.CategoryIds {
		category := &pb.Category{
			Id:   categoryID,
			Name: post.CategoryNames[i],
		}
		categories = append(categories, category)
	}

	tags := make([]*pb.Tag, 0, 5)
	for i, tagID := range post.TagIds {
		tag := &pb.Tag{
			Id:   tagID,
			Name: post.TagNames[i],
		}
		tags = append(tags, tag)
	}

	pbPost := &pb.Post{
		Id:         post.ID,
		Title:      post.Title,
		CoverImage: post.CoverImage,
		Content:    post.Content,
		Categories: categories,
		Tags:       tags,
		Featured:   post.Featured,
		Status:     post.Status,
	}

	return &pb.GetPostResponse{Post: pbPost}
}

func convertFeaturedPosts(posts []db.GetFeaturedPostsRow) *pb.GetFeaturedPostsResponse {
	if len(posts) == 0 {
		return &pb.GetFeaturedPostsResponse{}
	}

	rspPosts := []*pb.GetFeaturedPostsResponse_PostItem{}
	for _, post := range posts {
		author := &pb.UserItem{
			Id:       post.AuthorID,
			Username: post.Username,
			Avatar:   post.Avatar,
		}
		pbPost := &pb.GetFeaturedPostsResponse_PostItem{
			Id:           post.ID,
			Title:        post.Title,
			Author:       author,
			CoverImage:   post.CoverImage,
			ViewCount:    post.ViewCount,
			StarCount:    post.StarCount,
			CommentCount: post.CommentCount,
			PublishAt:    timestamppb.New(post.PublishAt),
		}
		rspPosts = append(rspPosts, pbPost)
	}
	return &pb.GetFeaturedPostsResponse{Posts: rspPosts}
}

func convertGetPosts(posts []db.GetPostsRow) *pb.GetPostsResponse {
	if len(posts) == 0 {
		return &pb.GetPostsResponse{}
	}

	rspPosts := []*pb.GetPostsResponse_PostItem{}
	for _, post := range posts {
		author := &pb.UserItem{
			Id:       post.AuthorID,
			Username: post.Username,
			Avatar:   post.Avatar,
		}

		tags := make([]*pb.Tag, 0, 5)
		for i, tagID := range post.TagIds {
			tag := &pb.Tag{
				Id:   tagID,
				Name: post.TagNames[i],
			}
			tags = append(tags, tag)
		}

		pbPost := &pb.GetPostsResponse_PostItem{
			Id:           post.ID,
			Title:        post.Title,
			Author:       author,
			CoverImage:   post.CoverImage,
			Tags:         tags,
			ViewCount:    post.ViewCount,
			StarCount:    post.StarCount,
			CommentCount: post.CommentCount,
			PublishAt:    timestamppb.New(post.PublishAt),
		}
		rspPosts = append(rspPosts, pbPost)
	}

	return &pb.GetPostsResponse{
		Total: posts[0].Total,
		Posts: rspPosts,
	}
}

func convertReadPost(post db.ReadPostRow) *pb.ReadPostResponse {
	author := &pb.UserInfo{
		Id:             post.AuthorID,
		Username:       post.Username,
		Avatar:         post.Avatar,
		Intro:          post.Intro,
		FollowerCount:  post.FollowerCount,
		FollowingCount: post.FollowingCount,
		Followed:       post.Followed.Valid,
	}

	categories := make([]*pb.Category, 0, 2)
	for i, categoryID := range post.CategoryIds {
		category := &pb.Category{
			Id:   categoryID,
			Name: post.CategoryNames[i],
		}
		categories = append(categories, category)
	}

	tags := make([]*pb.Tag, 0, 5)
	for i, tagID := range post.TagIds {
		tag := &pb.Tag{
			Id:   tagID,
			Name: post.TagNames[i],
		}
		tags = append(tags, tag)
	}

	pbPost := &pb.ReadPostResponse_Post{
		Id:         post.ID,
		Title:      post.Title,
		Author:     author,
		Content:    post.Content,
		Categories: categories,
		Tags:       tags,
		ViewCount:  post.ViewCount,
		StarCount:  post.StarCount,
		PublishAt:  timestamppb.New(post.PublishAt),
	}

	return &pb.ReadPostResponse{Post: pbPost}
}

func convertCreateComment(comment db.CreateCommentRow, user *db.User) *pb.CreateCommentResponse {
	var replyUser *pb.UserInfo
	if comment.ReplyUserID.Valid {
		replyUser = &pb.UserInfo{
			Id:             comment.ReplyUserID.Int64,
			Username:       comment.RUsername.String,
			Avatar:         comment.RAvatar.String,
			Intro:          comment.RIntro.String,
			FollowerCount:  comment.RFollowerCount,
			FollowingCount: comment.RFollowingCount,
			Followed:       comment.RFollowed.Valid,
		}
	}

	userInfo := &pb.UserInfo{
		Id:             user.ID,
		Username:       user.Username,
		Avatar:         user.Avatar,
		Intro:          user.Intro,
		FollowerCount:  comment.FollowerCount,
		FollowingCount: comment.FollowingCount,
	}

	return &pb.CreateCommentResponse{
		Id:        comment.ID,
		ParentId:  comment.ParentID.Int64,
		ReplyUser: replyUser,
		User:      userInfo,
		Content:   comment.Content,
		CreateAt:  timestamppb.New(comment.CreateAt),
	}
}

func convertListComments(comments []db.ListCommentsRow) *pb.ListCommentsResponse {
	if len(comments) == 0 {
		return &pb.ListCommentsResponse{}
	}

	rspComments := []*pb.Comment{}
	for _, comment := range comments {
		if comment.ParentID.Valid {
			continue
		}
		user := &pb.UserInfo{
			Id:             comment.UserID,
			Username:       comment.Username,
			Avatar:         comment.Avatar,
			Intro:          comment.Intro,
			FollowerCount:  comment.FollowerCount,
			FollowingCount: comment.FollowingCount,
			Followed:       comment.Followed.Valid,
		}
		pbComment := &pb.Comment{
			Id:         comment.ID,
			User:       user,
			Content:    comment.Content,
			StarCount:  comment.StarCount,
			ReplyCount: comment.ReplyCount,
			CreateAt:   timestamppb.New(comment.CreateAt),
		}
		rspComments = append(rspComments, pbComment)
	}

	for _, comment := range comments {
		if comment.ParentID.Valid {
			parentID := comment.ParentID.Int64

			var parent *pb.Comment
			for _, v := range rspComments {
				if v.Id == parentID {
					parent = v
					break
				}
			}
			user := &pb.UserInfo{
				Id:             comment.UserID,
				Username:       comment.Username,
				Avatar:         comment.Avatar,
				Intro:          comment.Intro,
				FollowerCount:  comment.FollowerCount,
				FollowingCount: comment.FollowingCount,
				Followed:       comment.Followed.Valid,
			}
			replyUser := &pb.UserInfo{
				Id:             comment.RUserID.Int64,
				Username:       comment.RUsername.String,
				Avatar:         comment.RAvatar.String,
				Intro:          comment.RIntro.String,
				FollowerCount:  comment.RFollowerCount,
				FollowingCount: comment.RFollowingCount,
				Followed:       comment.RFollowed.Valid,
			}
			pbReply := &pb.CommentReply{
				Id:        comment.ID,
				User:      user,
				ReplyUser: replyUser,
				Content:   comment.Content,
				StarCount: comment.StarCount,
				CreateAt:  timestamppb.New(comment.CreateAt),
			}
			parent.Replies = append(parent.Replies, pbReply)
		}
	}

	return &pb.ListCommentsResponse{
		Total:        comments[0].Total,
		CommentCount: comments[0].CommentCount,
		Comments:     rspComments,
	}
}

func convertListReplies(replies []db.ListRepliesRow) *pb.ListRepliesResponse {
	if len(replies) == 0 {
		return &pb.ListRepliesResponse{}
	}

	commentReplies := []*pb.CommentReply{}
	for _, reply := range replies {
		user := &pb.UserInfo{
			Id:             reply.UserID,
			Username:       reply.Username,
			Avatar:         reply.Avatar,
			Intro:          reply.Intro,
			FollowerCount:  reply.FollowerCount,
			FollowingCount: reply.FollowingCount,
			Followed:       reply.Followed.Valid,
		}
		replyUser := &pb.UserInfo{
			Id:             reply.RUserID.Int64,
			Username:       reply.RUsername.String,
			Avatar:         reply.RAvatar.String,
			Intro:          reply.RIntro.String,
			FollowerCount:  reply.RFollowerCount,
			FollowingCount: reply.RFollowingCount,
			Followed:       reply.RFollowed.Valid,
		}
		pbReply := &pb.CommentReply{
			Id:        reply.ID,
			User:      user,
			ReplyUser: replyUser,
			Content:   reply.Content,
			StarCount: reply.StarCount,
			CreateAt:  timestamppb.New(reply.CreateAt),
		}
		commentReplies = append(commentReplies, pbReply)
	}

	return &pb.ListRepliesResponse{
		Total:          replies[0].Total,
		CommentReplies: commentReplies,
	}
}
