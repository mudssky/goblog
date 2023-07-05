for (var i = 0; i < 100; i++) {
  // NumberInt可以将数据转换为int32 类型
  // timeStampNano= NumberInt((new Date()).getTime()*1000000)
  // NumberLong可以将数据转换为int64 类型
  // timeStampNano= NumberLong((new Date()).getTime()*1000000)
  _id = new ObjectId()
  date = new Date()
  newpost = {
    _id: _id,
    idhex: _id.toString().slice(10, -2),
    author: 'mudssky',
    title: 'test' + i,
    content: 'content' + i,
    creatat: date,
    lastmodified: date,
    categoryList: [],
    viewscounts: 0,
    commentCounts: 0,
  }
  db.post.InsertOne(c, newpost)
}
