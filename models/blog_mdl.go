package models

import (
	"../helpers/database"
	"errors"
	"github.com/ziutek/mymysql/mysql"
	//"log"
	"sort"
	"time"
)

type Posts []Post
type Post struct {
	ID           int
	Title        string
	Slug         string
	Content      string
	Published    time.Time
	Created      time.Time
	LastModified time.Time
	UserID       int
	MetaTitle    string
	MetaDesc     string
	Keywords     string
	Active       bool
	Categories   BlogCategories
	Comments     Comments
	Author       User
}

type BlogCategories []BlogCategory
type BlogCategory struct {
	ID     int
	Name   string
	Slug   string
	Active bool
}

type Comments struct {
	Approved   CommentList
	Unapproved CommentList
}
type CommentList []Comment
type Comment struct {
	ID       int
	PostID   int
	Name     string
	Email    string
	Comment  string
	Created  time.Time
	Approved bool
	Active   bool
}

func (p Post) GetAll() (Posts, error) {
	var posts Posts
	sel, err := database.GetStatement("GetAllPostsStmt")
	if err != nil {
		return posts, err
	}
	rows, res, err := sel.Exec()
	if err != nil {
		return posts, err
	}
	ch := make(chan Post)
	for _, row := range rows {
		go p.PopulatePost(row, res, ch)
	}
	for _, _ = range rows {
		posts = append(posts, <-ch)
	}
	return posts, nil
}

func (p Post) Get() (Post, error) {
	sel, err := database.GetStatement("GetPostStmt")
	if err != nil {
		return p, err
	}
	sel.Bind(p.ID)
	row, res, err := sel.ExecFirst()
	if err != nil {
		return p, err
	}
	ch := make(chan Post)
	go p.PopulatePost(row, res, ch)
	return <-ch, nil
}

func (c BlogCategory) GetAll() (BlogCategories, error) {
	var categories BlogCategories
	sel, err := database.GetStatement("GetAllBlogCategoriesStmt")
	if err != nil {
		return categories, err
	}
	rows, res, err := sel.Exec()
	if err != nil {
		return categories, err
	}
	ch := make(chan BlogCategory)
	for _, row := range rows {
		go c.PopulateCategory(row, res, ch)
	}
	for _, _ = range rows {
		categories = append(categories, <-ch)
	}
	return categories, nil
}

func (c BlogCategory) Get() (BlogCategory, error) {
	sel, err := database.GetStatement("GetBlogCategoryStmt")
	if err != nil {
		return BlogCategory{}, err
	}
	sel.Bind(c.ID)
	row, res, err := sel.ExecFirst()
	if err != nil {
		return BlogCategory{}, err
	}
	ch := make(chan BlogCategory)
	go c.PopulateCategory(row, res, ch)
	cat := <-ch
	return cat, nil
}

func (c *BlogCategory) Save() error {
	c.Slug = GenerateSlug(c.Name)
	if c.ID > 0 {
		// update
		upd, err := database.GetStatement("UpdateBlogCategoryStmt")
		if err != nil {
			return err
		}
		upd.Bind(c.Name, c.Slug, c.ID)
		_, _, err = upd.Exec()
		return err
	} else {
		// new
		ins, err := database.GetStatement("AddBlogCategoryStmt")
		if err != nil {
			return err
		}
		ins.Bind(c.Name, c.Slug)
		_, res, err := ins.Exec()
		if err != nil {
			return err
		}
		c.ID = int(res.InsertId())
	}
	return nil
}

func (c BlogCategory) Delete() bool {
	del, err := database.GetStatement("DeleteBlogCategoryStmt")
	if err != nil {
		return false
	}
	del.Bind(c.ID)
	_, _, err = del.Exec()
	if err != nil {
		return false
	}
	return true
}

func (c BlogCategory) AddToPost(postID int) {
	ins, err := database.GetStatement("AddPostCategoryStmt")
	if err != nil {
		return
	}
	ins.Raw.Reset()
	ins.Bind(postID, c.ID)
	ins.Exec()
}

func (p *Post) Save() error {
	if p.UserID == 0 {
		return errors.New("You must select an author")
	}
	if p.ID > 0 {
		// update post
		upd, err := database.GetStatement("UpdatePostStmt")
		if err != nil {
			return err
		}
		upd.Bind(p.Title, GenerateSlug(p.Title), p.Content, p.UserID, p.MetaTitle, p.MetaDesc, p.Keywords, p.ID)
		_, _, err = upd.Exec()
		if err != nil {
			return err
		}
		err = p.ClearCategories()
		if err != nil {
			return err
		}
	} else {
		// new post
		ins, err := database.GetStatement("AddPostStmt")
		if err != nil {
			return err
		}
		ins.Bind(p.Title, GenerateSlug(p.Title), p.Content, p.UserID, p.MetaTitle, p.MetaDesc, p.Keywords)
		_, res, err := ins.Exec()
		if err != nil {
			return err
		}
		p.ID = int(res.InsertId())
	}
	for i, _ := range p.Categories {
		go p.Categories[i].AddToPost(p.ID)
	}
	return nil
}

func (p Post) ClearCategories() error {
	del, err := database.GetStatement("ClearPostCategoriesStmt")
	if err != nil {
		return err
	}
	del.Bind(p.ID)
	_, _, err = del.Exec()
	return err
}

func (p *Post) Publish() error {
	upd, err := database.GetStatement("PublishPostStmt")
	if err != nil {
		return err
	}
	upd.Bind(p.ID)
	_, _, err = upd.Exec()
	return err
}

func (p *Post) UnPublish() error {
	upd, err := database.GetStatement("UnPublishPostStmt")
	if err != nil {
		return err
	}
	upd.Bind(p.ID)
	_, _, err = upd.Exec()
	p.Published = time.Time{}
	return err
}

func (p Post) Delete() bool {
	upd, err := database.GetStatement("DeletePostStmt")
	if err != nil {
		return false
	}
	upd.Bind(p.ID)
	_, _, err = upd.Exec()
	if err != nil {
		return false
	}
	return true
}

func (p Post) PopulatePost(row mysql.Row, res mysql.Result, ch chan Post) {
	post := Post{
		ID:           row.Int(res.Map("blogPostID")),
		Title:        row.Str(res.Map("post_title")),
		Slug:         row.Str(res.Map("slug")),
		Content:      row.Str(res.Map("post_text")),
		Published:    row.Time(res.Map("publishedDate"), UTC),
		Created:      row.Time(res.Map("createdDate"), UTC),
		LastModified: row.Time(res.Map("lastModified"), UTC),
		UserID:       row.Int(res.Map("userID")),
		MetaTitle:    row.Str(res.Map("meta_title")),
		MetaDesc:     row.Str(res.Map("meta_description")),
		Keywords:     row.Str(res.Map("keywords")),
		Active:       row.Bool(res.Map("active")),
	}
	catchan := make(chan BlogCategories)
	comchan := make(chan Comments)
	authchan := make(chan User)
	go func(ch chan User) {
		author, _ := GetUserByID(post.UserID)
		ch <- author
	}(authchan)
	go func(ch chan Comments) {
		comments, _ := post.GetComments()
		ch <- comments
	}(comchan)
	go func(ch chan BlogCategories) {
		categories, _ := post.GetCategories()
		ch <- categories
	}(catchan)

	post.Author = <-authchan
	post.Categories = <-catchan
	post.Comments = <-comchan
	ch <- post
}

func (p Post) GetCategories() (BlogCategories, error) {
	var categories BlogCategories
	sel, err := database.GetStatement("GetPostCategoriesStmt")
	if err != nil {
		return categories, err
	}
	sel.Bind(p.ID)
	rows, res, err := sel.Exec()
	if err != nil {
		return categories, err
	}
	ch := make(chan BlogCategory)
	for _, row := range rows {
		go BlogCategory{}.PopulateCategory(row, res, ch)
	}
	for _, _ = range rows {
		categories = append(categories, <-ch)
	}
	categories.Sort()
	return categories, nil
}

func (c BlogCategory) PopulateCategory(row mysql.Row, res mysql.Result, ch chan BlogCategory) {
	category := BlogCategory{
		ID:     row.Int(res.Map("blogCategoryID")),
		Name:   row.Str(res.Map("name")),
		Slug:   row.Str(res.Map("slug")),
		Active: row.Bool(res.Map("active")),
	}
	ch <- category
}

func (p Post) GetComments() (Comments, error) {
	var comments Comments
	var approved CommentList
	var unapproved CommentList
	sel, err := database.GetStatement("GetPostCommentsStmt")
	if err != nil {
		return comments, err
	}
	sel.Raw.Reset()
	sel.Bind(p.ID)
	rows, res, err := sel.Exec()
	if err != nil {
		return comments, err
	}
	ch := make(chan Comment)
	for i, _ := range rows {
		go Comment{}.PopulateComment(rows[i], res, ch)
	}
	for _, _ = range rows {
		c := <-ch
		if c.Approved {
			approved = append(approved, c)
		} else {
			unapproved = append(unapproved, c)
		}
	}
	approved.Sort()
	unapproved.Sort()
	comments.Approved = approved
	comments.Unapproved = unapproved
	return comments, nil
}

func (c Comment) PopulateComment(row mysql.Row, res mysql.Result, ch chan Comment) {
	comment := Comment{
		ID:       row.Int(res.Map("commentID")),
		PostID:   row.Int(res.Map("blogPostID")),
		Name:     row.Str(res.Map("name")),
		Email:    row.Str(res.Map("email")),
		Comment:  row.Str(res.Map("comment_text")),
		Created:  row.Time(res.Map("createdDate"), UTC),
		Approved: row.Bool(res.Map("approved")),
		Active:   row.Bool(res.Map("active")),
	}
	ch <- comment
}

func (c Comment) GetAll() (CommentList, error) {
	var comments CommentList
	sel, err := database.GetStatement("GetBlogCommentsStmt")
	if err != nil {
		return comments, err
	}
	rows, res, err := sel.Exec()
	if err != nil {
		return comments, err
	}
	ch := make(chan Comment)
	for _, row := range rows {
		go c.PopulateComment(row, res, ch)
	}
	for _, _ = range rows {
		comments = append(comments, <-ch)
	}
	return comments, nil
}

func (c Comment) Get() (Comment, error) {
	comment := Comment{}
	sel, err := database.GetStatement("GetBlogCommentStmt")
	if err != nil {
		return comment, err
	}
	sel.Bind(c.ID)
	row, res, err := sel.ExecFirst()
	if err != nil {
		return comment, err
	}
	ch := make(chan Comment)
	go c.PopulateComment(row, res, ch)
	return <-ch, nil
}

func (c Comment) Approve() bool {
	upd, err := database.GetStatement("ApproveBlogCommentStmt")
	if err != nil {
		return false
	}
	upd.Bind(c.ID)
	_, _, err = upd.Exec()
	if err != nil {
		return false
	}
	return true
}

func (c Comment) Delete() bool {
	upd, err := database.GetStatement("DeleteBlogCommentStmt")
	if err != nil {
		return false
	}
	upd.Bind(c.ID)
	_, _, err = upd.Exec()
	if err != nil {
		return false
	}
	return true
}

func (c BlogCategories) Len() int           { return len(c) }
func (c BlogCategories) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c BlogCategories) Less(i, j int) bool { return c[i].Name < (c[j].Name) }

func (c *BlogCategories) Sort() {
	sort.Sort(c)
}

func (c CommentList) Len() int           { return len(c) }
func (c CommentList) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c CommentList) Less(i, j int) bool { return c[i].Created.Before(c[j].Created) }

func (c *CommentList) Sort() {
	sort.Sort(c)
}

func (p Posts) Len() int           { return len(p) }
func (p Posts) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Posts) Less(i, j int) bool { return p[i].Published.Before(p[j].Published) }

func (p *Posts) Sort() {
	sort.Sort(p)
}
