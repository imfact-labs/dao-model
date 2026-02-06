package cmds

import (
	"context"

	ccmds "github.com/ProtoconNet/mitum-currency/v3/cmds"
	"github.com/ProtoconNet/mitum-dao/operation/dao"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

type PreSnapCommand struct {
	BaseCommand
	ccmds.OperationFlags
	Sender     ccmds.AddressFlag    `arg:"" name:"sender" help:"sender address" required:"true"`
	Contract   ccmds.AddressFlag    `arg:"" name:"contract" help:"contract address of credential" required:"true"`
	ProposalID string               `arg:"" name:"proposal-id" help:"proposal id" required:"true"`
	Currency   ccmds.CurrencyIDFlag `arg:"" name:"currency-id" help:"currency id" required:"true"`
	sender     base.Address
	contract   base.Address
}

func (cmd *PreSnapCommand) Run(pctx context.Context) error { // nolint:dupl
	if _, err := cmd.prepare(pctx); err != nil {
		return err
	}

	if err := cmd.parseFlags(); err != nil {
		return err
	}

	op, err := cmd.createOperation()
	if err != nil {
		return err
	}

	ccmds.PrettyPrint(cmd.Out, op)

	return nil
}

func (cmd *PreSnapCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	sender, err := cmd.Sender.Encode(cmd.Encoders.JSON())
	if err != nil {
		return errors.Wrapf(err, "invalid sender format, %q", cmd.Sender.String())
	}
	cmd.sender = sender

	contract, err := cmd.Contract.Encode(cmd.Encoders.JSON())
	if err != nil {
		return errors.Wrapf(err, "invalid contract account format, %q", cmd.Contract.String())
	}
	cmd.contract = contract

	return nil
}

func (cmd *PreSnapCommand) createOperation() (base.Operation, error) { // nolint:dupl}
	e := util.StringError("failed to create pre snap operation")

	fact := dao.NewPreSnapFact(
		[]byte(cmd.Token),
		cmd.sender,
		cmd.contract,
		cmd.ProposalID,
		cmd.Currency.CID,
	)

	op := dao.NewPreSnap(fact)
	err := op.HashSign(cmd.Privatekey, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, e.Wrap(err)
	}

	return op, nil
}
