package v6_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/cfnetworkingaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("remove-network-policy Command", func() {
	var (
		cmd             RemoveNetworkPolicyCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v6fakes.FakeRemoveNetworkPolicyActor
		binaryName      string
		executeErr      error
		srcApp          string
		destApp         string
		protocol        string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeRemoveNetworkPolicyActor)

		srcApp = "some-app"
		destApp = "some-other-app"
		protocol = "tcp"

		cmd = RemoveNetworkPolicyCommand{
			UI:             testUI,
			Config:         fakeConfig,
			SharedActor:    fakeSharedActor,
			Actor:          fakeActor,
			RequiredArgs:   flag.RemoveNetworkPolicyArgs{SourceApp: srcApp},
			DestinationApp: destApp,
			Protocol:       flag.NetworkProtocol{Protocol: protocol},
			Port:           flag.NetworkPort{StartPort: 8080, EndPort: 8081},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
		})

		It("outputs flavor text", func() {
			Expect(testUI.Out).To(Say(`Removing network policy for app %s in org some-org / space some-space as some-user\.\.\.`, srcApp))
		})

		When("the policy deletion is successful", func() {
			BeforeEach(func() {
				fakeActor.RemoveNetworkPolicyReturns(cfnetworkingaction.Warnings{"some-warning-1", "some-warning-2"}, nil)
			})

			It("displays OK", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(0))
				Expect(fakeActor.GetSpaceByNameAndOrganizationCallCount()).To(Equal(0))
				Expect(fakeActor.RemoveNetworkPolicyCallCount()).To(Equal(1))
				passedSrcSpaceGuid, passedSrcAppName, passedDestSpaceGuid, passedDestAppName, passedProtocol, passedStartPort, passedEndPort := fakeActor.RemoveNetworkPolicyArgsForCall(0)
				Expect(passedSrcSpaceGuid).To(Equal("some-space-guid"))
				Expect(passedSrcAppName).To(Equal("some-app"))
				Expect(passedDestSpaceGuid).To(Equal("some-space-guid"))
				Expect(passedDestAppName).To(Equal("some-other-app"))
				Expect(passedProtocol).To(Equal("tcp"))
				Expect(passedStartPort).To(Equal(8080))
				Expect(passedEndPort).To(Equal(8081))

				Expect(testUI.Out).To(Say(`Removing network policy for app %s in org some-org / space some-space as some-user\.\.\.`, srcApp))
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("when a valid org and space is provided", func() {
			BeforeEach(func() {
				cmd.DestinationOrg = "dest-org"
				cmd.DestinationSpace = "dest-space"
				fakeActor.GetOrganizationByNameReturns(v3action.Organization{GUID: "some-org-guid"}, v3action.Warnings{"some-warning-1"}, nil)
				fakeActor.GetSpaceByNameAndOrganizationReturns(v3action.Space{GUID: "some-dest-guid"}, v3action.Warnings{"some-warning-2"}, nil)
			})

			It("displays OK", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.RemoveNetworkPolicyCallCount()).To(Equal(1))
				passedSrcSpaceGuid, passedSrcAppName, passedDestSpaceGuid, passedDestAppName, passedProtocol, passedStartPort, passedEndPort := fakeActor.RemoveNetworkPolicyArgsForCall(0)
				Expect(passedSrcSpaceGuid).To(Equal("some-space-guid"))
				Expect(passedSrcAppName).To(Equal("some-app"))
				Expect(passedDestSpaceGuid).To(Equal("some-dest-guid"))
				Expect(passedDestAppName).To(Equal("some-other-app"))
				Expect(passedProtocol).To(Equal("tcp"))
				Expect(passedStartPort).To(Equal(8080))
				Expect(passedEndPort).To(Equal(8081))

				Expect(testUI.Out).To(Say(`Removing network policy for app %s in org some-org / space some-space as some-user\.\.\.`, srcApp))
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("a destination space but no destination org is specified", func() {
			BeforeEach(func() {
				cmd.DestinationSpace = "dest-space"
				fakeActor.GetSpaceByNameAndOrganizationReturns(v3action.Space{GUID: "some-dest-guid"}, v3action.Warnings{}, nil)
			})

			It("displays OK when no error occurs", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.RemoveNetworkPolicyCallCount()).To(Equal(1))
				Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(0))
				passedSrcSpaceGuid, passedSrcAppName, passedDestSpaceGuid, passedDestAppName, passedProtocol, passedStartPort, passedEndPort := fakeActor.RemoveNetworkPolicyArgsForCall(0)
				Expect(passedSrcSpaceGuid).To(Equal("some-space-guid"))
				Expect(passedSrcAppName).To(Equal("some-app"))
				Expect(passedDestSpaceGuid).To(Equal("some-dest-guid"))
				Expect(passedDestAppName).To(Equal("some-other-app"))
				Expect(passedProtocol).To(Equal("tcp"))
				Expect(passedStartPort).To(Equal(8080))
				Expect(passedEndPort).To(Equal(8081))

				Expect(testUI.Out).To(Say(`Removing network policy for app %s in org some-org / space some-space as some-user\.\.\.`, srcApp))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("a destination org is provided without a destination space", func() {
			BeforeEach(func() {
				cmd.DestinationOrg = "coolorg"
			})

			It("responds with an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.NetworkPolicyDestinationOrgWithoutSpaceError{}))
			})
		})

		When("an invalid org is provided", func() {
			BeforeEach(func() {
				cmd.DestinationOrg = "coolorg"
				cmd.DestinationSpace = "coolspace"
				fakeActor.GetOrganizationByNameReturns(v3action.Organization{}, v3action.Warnings{}, actionerror.OrganizationNotFoundError{Name: "coolorg"})
			})

			It("responds with an error", func() {
				passedOrgName := fakeActor.GetOrganizationByNameArgsForCall(0)
				Expect(passedOrgName).To(Equal("coolorg"))
				Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "coolorg"}))
			})
		})

		When("an invalid space is provided", func() {
			BeforeEach(func() {
				cmd.DestinationOrg = "coolorg"
				cmd.DestinationSpace = "coolspace"
				fakeActor.GetOrganizationByNameReturns(v3action.Organization{GUID: "some-org-guid"}, v3action.Warnings{}, nil)
				fakeActor.GetSpaceByNameAndOrganizationReturns(v3action.Space{}, v3action.Warnings{}, actionerror.SpaceNotFoundError{Name: "coolspace"})
			})

			It("responds with an error", func() {
				passedSpaceName, passedOrgGUID := fakeActor.GetSpaceByNameAndOrganizationArgsForCall(0)
				Expect(passedSpaceName).To(Equal("coolspace"))
				Expect(passedOrgGUID).To(Equal("some-org-guid"))
				Expect(executeErr).To(MatchError(actionerror.SpaceNotFoundError{Name: "coolspace"}))
			})
		})

		When("the policy does not exist", func() {
			BeforeEach(func() {
				fakeActor.RemoveNetworkPolicyReturns(cfnetworkingaction.Warnings{"some-warning-1", "some-warning-2"}, actionerror.PolicyDoesNotExistError{})
			})

			It("displays OK", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.RemoveNetworkPolicyCallCount()).To(Equal(1))
				passedSrcSpaceGuid, passedSrcAppName, passedDestSpaceGuid, passedDestAppName, passedProtocol, passedStartPort, passedEndPort := fakeActor.RemoveNetworkPolicyArgsForCall(0)
				Expect(passedSrcSpaceGuid).To(Equal("some-space-guid"))
				Expect(passedSrcAppName).To(Equal("some-app"))
				Expect(passedDestSpaceGuid).To(Equal("some-space-guid"))
				Expect(passedDestAppName).To(Equal("some-other-app"))
				Expect(passedProtocol).To(Equal("tcp"))
				Expect(passedStartPort).To(Equal(8080))
				Expect(passedEndPort).To(Equal(8081))

				Expect(testUI.Out).To(Say(`Removing network policy for app %s in org some-org / space some-space as some-user\.\.\.`, srcApp))
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
				Expect(testUI.Out).To(Say("Policy does not exist."))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("the policy deletion is not successful", func() {
			BeforeEach(func() {
				fakeActor.RemoveNetworkPolicyReturns(cfnetworkingaction.Warnings{"some-warning-1", "some-warning-2"}, actionerror.ApplicationNotFoundError{Name: srcApp})
			})

			It("does not display OK", func() {
				Expect(executeErr).To(MatchError(actionerror.ApplicationNotFoundError{Name: srcApp}))

				Expect(testUI.Out).To(Say(`Removing network policy for app %s in org some-org / space some-space as some-user\.\.\.`, srcApp))
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
				Expect(testUI.Out).ToNot(Say("OK"))
			})
		})
	})
})
