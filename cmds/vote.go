package cmds

import (
	"context"

	ccmds "github.com/imfact-labs/currency-model/app/cmds"
	"github.com/imfact-labs/dao-model/operation/dao"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/pkg/errors"
)

type VoteCommand struct {
	BaseCommand
	ccmds.OperationFlags
	Sender     ccmds.AddressFlag    `arg:"" name:"sender" help:"sender address" required:"true"`
	Contract   ccmds.AddressFlag    `arg:"" name:"contract" help:"contract address of credential" required:"true"`
	ProposalID string               `arg:"" name:"proposal-id" help:"proposal id" required:"true"`
	Vote       uint8                `arg:"" name:"vote" help:"vote" required:"true"`
	Currency   ccmds.CurrencyIDFlag `arg:"" name:"currency-id" help:"currency id" required:"true"`
	sender     base.Address
	contract   base.Address
}

func (cmd *VoteCommand) Run(pctx context.Context) error { // nolint:dupl
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

func (cmd *VoteCommand) parseFlags() error {
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

func (cmd *VoteCommand) createOperation() (base.Operation, error) { // nolint:dupl}
	e := util.StringError("failed to create vote operation")

	fact := dao.NewVoteFact(
		[]byte(cmd.Token),
		cmd.sender,
		cmd.contract,
		cmd.ProposalID,
		cmd.Vote,
		cmd.Currency.CID,
	)

	op := dao.NewVote(fact)
	err := op.HashSign(cmd.Privatekey, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, e.Wrap(err)
	}

	return op, nil
}
