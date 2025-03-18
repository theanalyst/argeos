Name: argeos
Version: 0.0.1 
# this is what comes after '-' in v0.0.1-{Release}
Release: 1%{?dist} 
Summary: The CERN 

Group: CERN-IT/SD
License: AGPLv3
ExclusiveArch: x86_64
Source0: %{name}-%{version}.tar.gz
Source1: systemd/argeos.service

BuildRequires: go-toolset, systemd

%description
Tasks
- Takes the status of the hostname from the PROBE and if down, will take diagnostics
- Is able to be remotely controlled to create the diagnostics reports

# do not strip the binary
%define __os_install_post %{nil}
# we do not provide debug packages, and the build checks that
%global debug_package %{nil}

%prep
%setup -q -n %{name}-%{version}

%install
install -d %{buildroot}%{_bindir}
install -d %{buildroot}%{_sysconfdir}/%{name}
install -m 755 %{name}            %{buildroot}%{_bindir}/
mkdir -p %{buildroot}/%{_unitdir}
install -D -m 0644 %{SOURCE1} %{buildroot}/%{_unitdir}/argeos.service

%clean
rm -rf %{buildroot}
rm -f %{name}/%{name}

%files
%defattr(-,root,root,-)
%{_bindir}/%{name}
/%{_unitdir}/argeos.service


%changelog
* Tue Mar 18 2025 Abhishek Lekshmanan  <abhishek.lekshmanan@cern.ch> 0.0.1-1
- Making the application alive
