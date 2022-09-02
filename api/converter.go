package api

import (
	"blog/server/db/sqlc"
	"blog/server/pb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertUser(user sqlc.User) *pb.User {
	return &pb.User{
		Id:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Info:     user.Info,
		Role:     user.Role,
	}
}

func convertListUsers(users []sqlc.ListUsersRow) *pb.ListUsersResponse {
	if len(users) == 0 {
		return &pb.ListUsersResponse{}
	}

	rspUsers := make([]*pb.ListUsersResponse_UserItem, 0, 5)
	for _, user := range users {
		pbUser := &pb.ListUsersResponse_UserItem{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Role:      user.Role,
			IsDeleted: user.IsDeleted,
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

func convertListSessions(sessions []sqlc.ListSessionsRow) *pb.ListSessionsResponse {
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

func convertListNotifs(notifs []sqlc.ListNotificationsRow, numRead int64) *pb.ListNotifsResponse {
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
		UnreadCount:   notifs[0].UnreadCount - numRead,
		Notifications: rspNotifs,
	}
}

func convertListFollowers(followers []sqlc.ListFollowersRow) *pb.ListFollowsResponse {
	if len(followers) == 0 {
		return &pb.ListFollowsResponse{}
	}

	rspUsers := make([]*pb.UserInfo, 0, 5)
	for _, follower := range followers {
		pbUserInfo := &pb.UserInfo{
			Id:         follower.UserID,
			Username:   follower.Username,
			Avatar:     follower.Avatar,
			Info:       follower.Info,
			IsFollowed: follower.Followed.Valid,
		}
		rspUsers = append(rspUsers, pbUserInfo)
	}

	return &pb.ListFollowsResponse{
		Total: followers[0].Total,
		Users: rspUsers,
	}
}

func convertListFollowings(followings []sqlc.ListFollowingsRow) *pb.ListFollowsResponse {
	if len(followings) == 0 {
		return &pb.ListFollowsResponse{}
	}

	rspUsers := make([]*pb.UserInfo, 0, 5)
	for _, follower := range followings {
		pbUserInfo := &pb.UserInfo{
			Id:         follower.UserID,
			Username:   follower.Username,
			Avatar:     follower.Avatar,
			Info:       follower.Info,
			IsFollowed: follower.Followed.Valid,
		}
		rspUsers = append(rspUsers, pbUserInfo)
	}

	return &pb.ListFollowsResponse{
		Total: followings[0].Total,
		Users: rspUsers,
	}
}

func convertCategory(category sqlc.Category) *pb.Category {
	return &pb.Category{
		Id:   category.ID,
		Name: category.Name,
	}
}

func convertCategories(categories []sqlc.Category) []*pb.Category {
	rsp := []*pb.Category{}
	for _, category := range categories {
		rsp = append(rsp, convertCategory(category))
	}
	return rsp
}

func convertListCategories(categories []sqlc.ListCategoriesRow) *pb.ListCategoriesResponse {
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

func convertTag(tag sqlc.Tag) *pb.Tag {
	return &pb.Tag{
		Id:   tag.ID,
		Name: tag.Name,
	}
}

func convertTags(tags []sqlc.Tag) []*pb.Tag {
	rsp := []*pb.Tag{}
	for _, tag := range tags {
		rsp = append(rsp, convertTag(tag))
	}
	return rsp
}

func convertListTags(tags []sqlc.ListTagsRow) *pb.ListTagsResponse {
	if len(tags) == 0 {
		return &pb.ListTagsResponse{}
	}

	rspTags := []*pb.ListTagsResponse_TagItem{}
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

func convertPost(post sqlc.Post, categories []sqlc.Category, tags []sqlc.Tag) *pb.Post {
	return &pb.Post{
		Id:         post.ID,
		Title:      post.Title,
		Abstract:   post.Abstract,
		CoverImage: post.CoverImage,
		Content:    post.Content,
		Categories: convertCategories(categories),
		Tags:       convertTags(tags),
		IsFeatured: post.IsFeatured,
		Status:     post.Status,
	}
}

func convertListPosts(posts []sqlc.ListPostsRow) *pb.ListPostsResponse {
	if len(posts) == 0 {
		return &pb.ListPostsResponse{}
	}

	rspPosts := []*pb.ListPostsResponse_PostItem{}
	for _, post := range posts {
		pbPost := &pb.ListPostsResponse_PostItem{
			Id:        post.ID,
			Title:     post.Title,
			Status:    post.Status,
			UpdateAt:  timestamppb.New(post.UpdateAt),
			PublishAt: timestamppb.New(post.PublishAt),
		}
		rspPosts = append(rspPosts, pbPost)
	}
	return &pb.ListPostsResponse{
		Total: posts[0].Total,
		Posts: rspPosts,
	}
}

func convertGetPosts(posts []sqlc.GetPostsRow) *pb.GetPostsResponse {
	if len(posts) == 0 {
		return &pb.GetPostsResponse{}
	}

	rspPosts := []*pb.GetPostsResponse_PostItem{}
	for _, post := range posts {
		author := &pb.User{
			Id:       post.AuthorID,
			Username: post.Username,
			Email:    post.Email,
			Avatar:   post.Avatar,
		}

		categories := []*pb.Category{}
		for i := 0; i < len(post.CategoryIds); i++ {
			category := &pb.Category{
				Id:   post.CategoryIds[i],
				Name: post.CategoryNames[i],
			}
			categories = append(categories, category)
		}

		pbPost := &pb.GetPostsResponse_PostItem{
			Id:         post.ID,
			Title:      post.Title,
			Author:     author,
			ViewCount:  post.ViewCount,
			Categories: categories,
			Status:     post.Status,
			IsFeatured: post.IsFeatured,
			UpdateAt:   timestamppb.New(post.UpdateAt),
			PublishAt:  timestamppb.New(post.PublishAt),
		}
		rspPosts = append(rspPosts, pbPost)
	}
	return &pb.GetPostsResponse{
		Total: posts[0].Total,
		Posts: rspPosts,
	}
}

func convertFetchPosts(posts []sqlc.FetchPostsRow) *pb.FetchPostsResponse {
	if len(posts) == 0 {
		return &pb.FetchPostsResponse{}
	}

	rspPosts := []*pb.FetchPostsResponse_PostItem{}
	for _, post := range posts {
		author := &pb.UserInfo{
			Id:             post.AuthorID,
			Username:       post.Username,
			Info:           post.Info,
			Avatar:         post.Avatar,
			FollowerCount:  post.FollowerCount,
			FollowingCount: post.FollowingCount,
			IsFollowed:     post.Followed.Valid,
		}

		tags := []*pb.Tag{}
		for i := 0; i < len(post.TagIds); i++ {
			tag := &pb.Tag{
				Id:   post.TagIds[i],
				Name: post.TagNames[i],
			}
			tags = append(tags, tag)
		}

		pbPost := &pb.FetchPostsResponse_PostItem{
			Id:           post.ID,
			Title:        post.Title,
			Author:       author,
			Abstract:     post.Abstract,
			CoverImage:   post.CoverImage,
			Tags:         tags,
			ViewCount:    post.ViewCount,
			StarCount:    post.StarCount,
			CommentCount: post.CommentCount,
			PublishAt:    timestamppb.New(post.PublishAt),
		}
		rspPosts = append(rspPosts, pbPost)
	}
	return &pb.FetchPostsResponse{
		Total: posts[0].Total,
		Posts: rspPosts,
	}
}

func convertReviewPost(post sqlc.ReviewPostRow) *pb.ReviewPostResponse {
	tags := []*pb.Tag{}
	for i := 0; i < len(post.TagIds); i++ {
		tag := &pb.Tag{
			Id:   post.TagIds[i],
			Name: post.TagNames[i],
		}
		tags = append(tags, tag)
	}

	categories := []*pb.Category{}
	for i := 0; i < len(post.CategoryIds); i++ {
		category := &pb.Category{
			Id:   post.CategoryIds[i],
			Name: post.CategoryNames[i],
		}
		categories = append(categories, category)
	}

	pbPost := &pb.Post{
		Id:         post.ID,
		Title:      post.Title,
		Abstract:   post.Abstract,
		CoverImage: post.CoverImage,
		Content:    post.Content,
		Categories: categories,
		Tags:       tags,
		IsFeatured: post.IsFeatured,
		Status:     post.Status,
	}

	return &pb.ReviewPostResponse{Post: pbPost}
}

func convertReadPost(post sqlc.ReadPostRow) *pb.ReadPostResponse {
	author := &pb.UserInfo{
		Id:             post.AuthorID,
		Username:       post.Username,
		Avatar:         post.Avatar,
		Info:           post.Info,
		FollowerCount:  post.FollowerCount,
		FollowingCount: post.FollowingCount,
		IsFollowed:     post.Followed.Valid,
	}

	tags := []*pb.Tag{}
	for i := 0; i < len(post.TagIds); i++ {
		tag := &pb.Tag{
			Id:   post.TagIds[i],
			Name: post.TagNames[i],
		}
		tags = append(tags, tag)
	}

	categories := []*pb.Category{}
	for i := 0; i < len(post.CategoryIds); i++ {
		category := &pb.Category{
			Id:   post.CategoryIds[i],
			Name: post.CategoryNames[i],
		}
		categories = append(categories, category)
	}

	pbPost := &pb.ReadPostResponse_Post{
		Id:         post.ID,
		Title:      post.Title,
		Author:     author,
		CoverImage: post.CoverImage,
		Content:    post.Content,
		Categories: categories,
		Tags:       tags,
		ViewCount:  post.ViewCount,
		StarCount:  post.StarCount,
		PublishAt:  timestamppb.New(post.PublishAt),
	}

	return &pb.ReadPostResponse{Post: pbPost}
}

func convertCreateComment(comment sqlc.CreateCommentRow, user sqlc.User) *pb.CreateCommentResponse {
	var replyUser *pb.UserInfo
	if comment.ReplyUserID.Valid {
		replyUser = &pb.UserInfo{
			Id:             comment.ReplyUserID.Int64,
			Username:       comment.RUsername.String,
			Avatar:         comment.RAvatar.String,
			Info:           comment.RInfo.String,
			FollowerCount:  comment.RFollowerCount,
			FollowingCount: comment.RFollowingCount,
			IsFollowed:     comment.RFollowed.Valid,
		}
	}

	userInfo := &pb.UserInfo{
		Id:             user.ID,
		Username:       user.Username,
		Avatar:         user.Avatar,
		Info:           user.Info,
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

func convertListComments(comments []sqlc.ListCommentsRow) *pb.ListCommentsResponse {
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
			Info:           comment.Info,
			FollowerCount:  comment.FollowerCount,
			FollowingCount: comment.FollowingCount,
			IsFollowed:     comment.Followed.Valid,
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
				Info:           comment.Info,
				FollowerCount:  comment.FollowerCount,
				FollowingCount: comment.FollowingCount,
				IsFollowed:     comment.Followed.Valid,
			}
			replyUser := &pb.UserInfo{
				Id:             comment.RUserID.Int64,
				Username:       comment.RUsername.String,
				Avatar:         comment.RAvatar.String,
				Info:           comment.RInfo.String,
				FollowerCount:  comment.RFollowerCount,
				FollowingCount: comment.RFollowingCount,
				IsFollowed:     comment.RFollowed.Valid,
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

func convertListReplies(replies []sqlc.ListRepliesRow) *pb.ListRepliesResponse {
	if len(replies) == 0 {
		return &pb.ListRepliesResponse{}
	}

	commentReplies := []*pb.CommentReply{}
	for _, reply := range replies {
		user := &pb.UserInfo{
			Id:             reply.UserID,
			Username:       reply.Username,
			Avatar:         reply.Avatar,
			Info:           reply.Info,
			FollowerCount:  reply.FollowerCount,
			FollowingCount: reply.FollowingCount,
			IsFollowed:     reply.Followed.Valid,
		}
		replyUser := &pb.UserInfo{
			Id:             reply.RUserID.Int64,
			Username:       reply.RUsername.String,
			Avatar:         reply.RAvatar.String,
			Info:           reply.RInfo.String,
			FollowerCount:  reply.RFollowerCount,
			FollowingCount: reply.RFollowingCount,
			IsFollowed:     reply.RFollowed.Valid,
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
