-- name: CardBelongsToBoard :one
-- Check if a card with given card_id belongs to a board with given board_id
SELECT EXISTS (
  SELECT 1
  FROM kanban_cards
  WHERE card_id = $1 AND board_id = $2
);

-- name: CommentBelongsToCard :one
-- Check if a comment with given comment_id belongs to a card with given card_id
SELECT EXISTS (
  SELECT 1
  FROM card_comments
  WHERE comment_id = $1 AND card_id = $2
);

-- name: BoardCardCommentRelationshipExists :one
-- Check if a comment, card, and board are all related
SELECT EXISTS (
  SELECT 1
  FROM card_comments cmt
  JOIN kanban_cards crd ON cmt.card_id = crd.card_id
  JOIN kanban_boards brd ON crd.board_id = brd.board_id
  WHERE cmt.comment_id = $1 AND crd.card_id = $2 AND brd.board_id = $3
);