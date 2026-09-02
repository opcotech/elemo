/*
 * Elemo license policy for ORT.
 *
 * FSL-1.1-ALv2 is not in the OSADL matrix. Dependency compatibility is checked
 * against Apache-2.0, the irrevocable future license of FSL-1.1-ALv2.
 */

import org.ossreviewtoolkit.evaluator.osadl.Compatibility
import org.ossreviewtoolkit.evaluator.osadl.CompatibilityMatrix

fun getLicensesForCategory(category: String): Set<SpdxSingleLicenseExpression> =
    licenseClassifications.licensesByCategory[category].orEmpty()

val commercialLicenses = getLicensesForCategory("commercial")
val copyleftLicenses = getLicensesForCategory("copyleft")
val copyleftLimitedLicenses = getLicensesForCategory("copyleft-limited")
val freeRestrictedLicenses = getLicensesForCategory("free-restricted")
val genericLicenses = getLicensesForCategory("generic")
val patentLicenses = getLicensesForCategory("patent-license")
val permissiveLicenses = getLicensesForCategory("permissive")
val proprietaryFreeLicenses = getLicensesForCategory("proprietary-free")
val publicDomainLicenses = getLicensesForCategory("public-domain")
val sourceAvailableLicenses = getLicensesForCategory("source-available")
val unknownLicenses = getLicensesForCategory("unknown")
val unstatedLicenses = getLicensesForCategory("unstated-license")
val claLicenses = getLicensesForCategory("cla")

val ignoredLicenses = listOf(
    "LicenseRef-scancode-generic-cla",
    "LicenseRef-scancode-other-permissive",
    "LicenseRef-scancode-public-domain",
    "LicenseRef-scancode-public-domain-disclaimer",
    "LicenseRef-scancode-us-govt-public-domain",
    "LicenseRef-scancode-warranty-disclaimer",
    "LicenseRef-scancode-license-file-reference",
    "LicenseRef-scancode-see-license",
    "LicenseRef-scancode-unknown-license-reference"
).map { SpdxSingleLicenseExpression.parse(it) }.toSet()

val handledLicenses = listOf(
    claLicenses,
    commercialLicenses,
    copyleftLicenses,
    copyleftLimitedLicenses,
    freeRestrictedLicenses,
    genericLicenses,
    patentLicenses,
    permissiveLicenses,
    proprietaryFreeLicenses,
    publicDomainLicenses,
    sourceAvailableLicenses,
    unknownLicenses,
    unstatedLicenses
).flatten().toSet()

val outboundLicense = "Apache-2.0"

fun PackageRule.howToFixDefault() = """
    Replace the dependency, add a package curation, or document a resolution in `.ort.yml`.
""".trimIndent()

fun isExceptionWithoutLicense(license: SpdxSingleLicenseExpression) =
    "-exception" in license.toString() && " WITH " !in license.toString()

fun PackageRule.LicenseRule.isHandled() =
    object : RuleMatcher {
        override val description = "isHandled($license)"

        override fun matches() =
            license in handledLicenses && !isExceptionWithoutLicense(license)
    }

fun PackageRule.LicenseRule.isCommercial() =
    object : RuleMatcher {
        override val description = "isCommercial($license)"

        override fun matches() = license in commercialLicenses
    }

fun PackageRule.LicenseRule.isCopyleft() =
    object : RuleMatcher {
        override val description = "isCopyleft($license)"

        override fun matches() = license in copyleftLicenses
    }

fun PackageRule.LicenseRule.isCopyleftLimited() =
    object : RuleMatcher {
        override val description = "isCopyleftLimited($license)"

        override fun matches() = license in copyleftLimitedLicenses
    }

fun PackageRule.LicenseRule.isIgnored() =
    object : RuleMatcher {
        override val description = "isIgnored($license)"

        override fun matches() = license in ignoredLicenses
    }

fun PackageRule.LicenseRule.isProprietaryFree() =
    object : RuleMatcher {
        override val description = "isProprietaryFree($license)"

        override fun matches() = license in proprietaryFreeLicenses
    }

fun PackageRule.LicenseRule.isUnknown() =
    object : RuleMatcher {
        override val description = "isUnknown($license)"

        override fun matches() = license in unknownLicenses
    }

fun PackageRule.LicenseRule.isUnstated() =
    object : RuleMatcher {
        override val description = "isUnstated($license)"

        override fun matches() = license in unstatedLicenses
    }

fun RuleSet.unhandledLicenseRule() = packageRule("UNHANDLED_LICENSE") {
    require {
        -isExcluded()
    }

    licenseRule("UNHANDLED_LICENSE", LicenseView.CONCLUDED_OR_DECLARED_AND_DETECTED) {
        require {
            -isExcluded()
            -isHandled()
            -isIgnored()
        }

        error(
            "The license $license is not covered by Elemo policy rules. " +
                "It was ${licenseSource.name.lowercase()} in package '${pkg.metadata.id.toCoordinates()}'.",
            howToFixDefault()
        )
    }
}

fun RuleSet.unmappedDeclaredLicenseRule() = packageRule("UNMAPPED_DECLARED_LICENSE") {
    require {
        -isExcluded()
    }

    resolvedLicenseInfo.licenseInfo.declaredLicenseInfo.processed.unmapped.forEach { unmappedLicense ->
        warning(
            "The declared license '$unmappedLicense' could not be mapped to a valid SPDX expression in package " +
                "'${pkg.metadata.id.toCoordinates()}'.",
            howToFixDefault()
        )
    }
}

fun RuleSet.noLicenseInDependencyRule() = packageRule("NO_LICENSE_IN_DEPENDENCY") {
    require {
        -isProject()
        -isExcluded()
        -hasLicense()
    }

    error(
        "No license information is available for dependency '${pkg.metadata.id.toCoordinates()}'.",
        "If the dependency is unlicensed it must not be used. Otherwise conclude the license with a package curation."
    )
}

fun RuleSet.copyleftInSourceRule() = packageRule("COPYLEFT_IN_SOURCE") {
    require {
        +isProject()
        -isExcluded()
    }

    licenseRule("COPYLEFT_IN_SOURCE", LicenseView.CONCLUDED_OR_DECLARED_AND_DETECTED) {
        require {
            -isExcluded()
            +isCopyleft()
        }

        error(
            "The copyleft license $license was ${licenseSource.name.lowercase()} in project " +
                "'${pkg.metadata.id.toCoordinates()}'.",
            howToFixDefault()
        )
    }
}

fun RuleSet.copyleftLimitedInSourceRule() = packageRule("COPYLEFT_LIMITED_IN_SOURCE") {
    require {
        +isProject()
        -isExcluded()
    }

    licenseRule("COPYLEFT_LIMITED_IN_SOURCE", LicenseView.CONCLUDED_OR_DECLARED_AND_DETECTED) {
        require {
            -isExcluded()
            +isCopyleftLimited()
        }

        error(
            "The copyleft-limited license $license was ${licenseSource.name.lowercase()} in project " +
                "'${pkg.metadata.id.toCoordinates()}'.",
            howToFixDefault()
        )
    }
}

fun RuleSet.commercialInDependencyRule() = packageRule("COMMERCIAL_IN_DEPENDENCY") {
    require {
        -isProject()
        -isExcluded()
    }

    licenseRule("COMMERCIAL_IN_DEPENDENCY", LicenseView.CONCLUDED_OR_DECLARED_AND_DETECTED) {
        require {
            +isCommercial()
            -isExcluded()
        }

        error(
            "The dependency '${pkg.metadata.id.toCoordinates()}' uses commercial license $license.",
            howToFixDefault()
        )
    }
}

fun RuleSet.proprietaryFreeInDependencyRule() = packageRule("PROPRIETARY_FREE_IN_DEPENDENCY") {
    require {
        -isProject()
        -isExcluded()
    }

    licenseRule("PROPRIETARY_FREE_IN_DEPENDENCY", LicenseView.CONCLUDED_OR_DECLARED_AND_DETECTED) {
        require {
            +isProprietaryFree()
            -isExcluded()
        }

        error(
            "The dependency '${pkg.metadata.id.toCoordinates()}' uses proprietary-free license $license.",
            howToFixDefault()
        )
    }
}

fun RuleSet.unknownInDependencyRule() = packageRule("UNKNOWN_IN_DEPENDENCY") {
    require {
        -isProject()
        -isExcluded()
    }

    licenseRule("UNKNOWN_IN_DEPENDENCY", LicenseView.CONCLUDED_OR_DECLARED_AND_DETECTED) {
        require {
            +isUnknown()
            -isIgnored()
            -isExcluded()
        }

        error(
            "The dependency '${pkg.metadata.id.toCoordinates()}' uses unknown license $license.",
            howToFixDefault()
        )
    }
}

fun RuleSet.unstatedInDependencyRule() = packageRule("UNSTATED_IN_DEPENDENCY") {
    require {
        -isProject()
        -isExcluded()
    }

    licenseRule("UNSTATED_IN_DEPENDENCY", LicenseView.CONCLUDED_OR_DECLARED_AND_DETECTED) {
        require {
            +isUnstated()
            -isExcluded()
        }

        error(
            "The dependency '${pkg.metadata.id.toCoordinates()}' uses unstated license $license.",
            howToFixDefault()
        )
    }
}

fun RuleSet.osadlCompatibilityRule() = dependencyRule("OSADL_MATRIX_COMPATIBILITY") {
    require {
        -isExcluded()
    }

    licenseRule("OSADL_PROJECT_LICENSE_COMPATIBILITY", LicenseView.CONCLUDED_OR_DECLARED_AND_DETECTED) {
        require {
            -isExcluded()
        }

        val compatibilityInfo = CompatibilityMatrix.getCompatibilityInfo(
            outboundLicense,
            license.simpleLicense()
        )

        if (compatibilityInfo.compatibility !in Compatibility.COMPATIBLE_VALUES) {
            val inbound = license.toString()
            val depCoords = dependency.id.toCoordinates()

            when (compatibilityInfo.compatibility) {
                Compatibility.CONTEXTUAL -> warning(
                    "Whether Apache-2.0 (FSL-1.1-ALv2 future license) is compatible with inbound license $inbound " +
                        "of '$depCoords' depends on the context. ${compatibilityInfo.explanation}",
                    "Get legal advice and add a rule-violation resolution in `.ort.yml` if the use is approved."
                )

                Compatibility.UNKNOWN -> warning(
                    "It is unknown whether Apache-2.0 (FSL-1.1-ALv2 future license) is compatible with inbound " +
                        "license $inbound of '$depCoords'. ${compatibilityInfo.explanation}",
                    "Get legal advice and add a rule-violation resolution in `.ort.yml` if the use is approved."
                )

                else -> error(
                    "Apache-2.0 (FSL-1.1-ALv2 future license) is incompatible with inbound license $inbound of " +
                        "'$depCoords'. ${compatibilityInfo.explanation}",
                    "Remove the dependency or replace it with a license-compatible alternative."
                )
            }
        }
    }
}

val ruleSet = ruleSet(ortResult, licenseInfoResolver) {
    unhandledLicenseRule()
    unmappedDeclaredLicenseRule()
    noLicenseInDependencyRule()
    commercialInDependencyRule()
    proprietaryFreeInDependencyRule()
    unknownInDependencyRule()
    unstatedInDependencyRule()
    copyleftInSourceRule()
    copyleftLimitedInSourceRule()
    osadlCompatibilityRule()
}

ruleViolations += ruleSet.violations
