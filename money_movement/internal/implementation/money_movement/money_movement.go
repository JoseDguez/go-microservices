package money_movement

import (
	"context"
	"database/sql"
	"errors"
	"github.com/JoseDguez/go-microservices/internal/producer"
	pb "github.com/JoseDguez/go-microservices/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	insertTransactionQuery = "INSERT INTO transactions (pid, src_user_id, dst_user_id, src_wallet_id, dst_wallet_id, src_account_id, dst_account_id, src_account_type, dst_account_type, final_dst_wallet_id, amount) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	selectTransactionQuery = "SELECT id, pid, src_user_id, dst_user_id, src_wallet_id, dst_wallet_id, src_account_id, dst_account_id, src_account_type, dst_account_type, final_dst_wallet_id, amount FROM transactions WHERE pid = ?"
)

type Implementation struct {
	db *sql.DB
	pb.UnimplementedMoneyMovementServiceServer
}

func NewMoneyMovementImplementation(db *sql.DB) *Implementation {
	return &Implementation{db: db}
}

func (this *Implementation) Authorize(ctx context.Context, authorizePayload *pb.AuthorizePayload) (*pb.AuthorizeResponse, error) {
	if authorizePayload.GetCurrency() != "USD" {
		return nil, status.Error(codes.InvalidArgument, "only USD currency is supported")
	}

	tx, err := this.db.Begin()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	merchantWallet, err := fetchWallet(tx, authorizePayload.GetMerchantWalletUserId())
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return nil, status.Error(codes.Internal, rollbackErr.Error())
		}
		return nil, err
	}

	customerWallet, err := fetchWallet(tx, authorizePayload.GetCustomerWalletUserId())
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return nil, status.Error(codes.Internal, rollbackErr.Error())
		}
		return nil, err
	}

	srcAccount, err := fetchAccount(tx, customerWallet.ID, "DEFAULT")
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return nil, status.Error(codes.Internal, rollbackErr.Error())
		}
		return nil, err
	}

	dstAccount, err := fetchAccount(tx, customerWallet.ID, "PAYMENT")
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return nil, status.Error(codes.Internal, rollbackErr.Error())
		}
		return nil, err
	}

	err = transfer(tx, srcAccount, dstAccount, authorizePayload.GetCents())
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return nil, status.Error(codes.Internal, rollbackErr.Error())
		}
		return nil, err
	}

	pid := uuid.NewString()
	err = createTransaction(tx, pid, srcAccount, dstAccount, customerWallet, customerWallet, merchantWallet, authorizePayload.GetCents())
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return nil, status.Error(codes.Internal, rollbackErr.Error())
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AuthorizeResponse{Pid: pid}, nil
}

func (this *Implementation) Capture(ctx context.Context, capturePayload *pb.CapturePayload) (*emptypb.Empty, error) {
	tx, err := this.db.Begin()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	authorizedTransaction, err := fetchTransaction(tx, capturePayload.GetPid())
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return nil, status.Error(codes.Internal, rollbackErr.Error())
		}
		return nil, err
	}

	srcAccount, err := fetchAccount(tx, authorizedTransaction.dstAccountWalletID, "PAYMENT")
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return nil, status.Error(codes.Internal, rollbackErr.Error())
		}
		return nil, err
	}

	dstMerchantAccount, err := fetchAccount(tx, authorizedTransaction.finalDstMerchantWalletID, "INCOMING")
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return nil, status.Error(codes.Internal, rollbackErr.Error())
		}
		return nil, err
	}

	err = transfer(tx, srcAccount, dstMerchantAccount, authorizedTransaction.amount)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return nil, status.Error(codes.Internal, rollbackErr.Error())
		}
		return nil, err
	}

	customerWallet, err := fetchWallet(tx, authorizedTransaction.srcUserID)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return nil, status.Error(codes.Internal, rollbackErr.Error())
		}
		return nil, err
	}

	merchantWallet, err := findWallet(tx, authorizedTransaction.finalDstMerchantWalletID)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return nil, status.Error(codes.Internal, rollbackErr.Error())
		}
		return nil, err
	}

	err = createTransaction(tx, authorizedTransaction.pid, srcAccount, dstMerchantAccount, customerWallet, merchantWallet, merchantWallet, authorizedTransaction.amount)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return nil, status.Error(codes.Internal, rollbackErr.Error())
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	producer.SendCaptureMessage(authorizedTransaction.pid, authorizedTransaction.srcUserID, authorizedTransaction.amount)

	return &emptypb.Empty{}, nil
}

func fetchWallet(tx *sql.Tx, userID string) (wallet, error) {
	var w wallet

	stmt, err := tx.Prepare("SELECT id, user_id, wallet_type FROM wallets WHERE user_id = ?")
	if err != nil {
		return w, status.Error(codes.Internal, err.Error())
	}

	err = stmt.QueryRow(userID).Scan(&w.ID, &w.userID, &w.walletType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return w, status.Error(codes.InvalidArgument, "wallet not found")
		}
		return w, status.Error(codes.Internal, err.Error())
	}

	return w, nil
}

func findWallet(tx *sql.Tx, walletID int32) (wallet, error) {
	var w wallet

	stmt, err := tx.Prepare("SELECT id, user_id, wallet_type FROM wallets WHERE id = ?")
	if err != nil {
		return w, status.Error(codes.Internal, err.Error())
	}

	err = stmt.QueryRow(walletID).Scan(&w.ID, &w.userID, &w.walletType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return w, status.Error(codes.InvalidArgument, "wallet not found")
		}
		return w, status.Error(codes.Internal, err.Error())
	}

	return w, nil
}

func fetchAccount(tx *sql.Tx, walletID int32, accountType string) (account, error) {
	var a account

	stmt, err := tx.Prepare("SELECT id, cents, account_type, wallet_id FROM accounts WHERE wallet_id = ? AND account_type = ?")
	if err != nil {
		return a, status.Error(codes.Internal, err.Error())
	}

	err = stmt.QueryRow(walletID, accountType).Scan(&a.ID, &a.cents, &a.accountType, &a.walletID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return a, status.Error(codes.InvalidArgument, "account not found")
		}
		return a, status.Error(codes.Internal, err.Error())
	}

	return a, nil
}

func transfer(tx *sql.Tx, srcAccount account, dstAccount account, amount int64) error {
	if srcAccount.cents < amount {
		return status.Error(codes.InvalidArgument, "insufficient funds")
	}

	stmt, err := tx.Prepare("UPDATE accounts SET cents = ? WHERE id = ?")
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	_, err = stmt.Exec(srcAccount.cents-amount, srcAccount.ID)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	stmt, err = tx.Prepare("UPDATE accounts SET cents = ? WHERE id = ?")
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	_, err = stmt.Exec(dstAccount.cents+amount, dstAccount.ID)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func createTransaction(tx *sql.Tx, pid string, srcAccount account, dstAccount account, srcWallet wallet, dstWallet wallet, finalDstWallet wallet, amount int64) error {
	stmt, err := tx.Prepare(insertTransactionQuery)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	_, err = stmt.Exec(
		pid,
		srcWallet.userID,
		dstWallet.userID,
		srcAccount.walletID,
		dstAccount.walletID,
		srcAccount.ID,
		dstAccount.ID,
		srcAccount.accountType,
		dstAccount.accountType,
		finalDstWallet.ID,
		amount)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func fetchTransaction(tx *sql.Tx, pid string) (transaction, error) {
	var t transaction

	stmt, err := tx.Prepare(selectTransactionQuery)
	if err != nil {
		return t, status.Error(codes.Internal, err.Error())
	}

	err = stmt.QueryRow(pid).Scan(&t.ID, &t.pid, &t.srcUserID, &t.dstUserID, &t.srcAccountWalletID, &t.dstAccountWalletID, &t.srcAccountID, &t.dstAccountID, &t.srcAccountType, &t.dstAccountType, &t.finalDstMerchantWalletID, &t.amount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return t, status.Error(codes.InvalidArgument, "transaction not found")
		}
		return t, status.Error(codes.Internal, err.Error())
	}

	return t, nil
}
